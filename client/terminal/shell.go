package terminal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"kube-terminal/model/request"
	"log"
	"net/http"
	"sync"
	"time"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"
const SessionTerminalStoreTime = 5 // session timeout (minute)

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// TerminalSession implements PtyHandler (using a SockJS connection)
type TerminalSession struct {
	Id            string
	Bound         chan error
	sockJSSession sockjs.Session
	SizeChan      chan remotecommand.TerminalSize
	doneChan      chan struct{}
	TimeOut       time.Time
	RequestInfo   request.TerminalRequest
}

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	Op, Data, SessionID string
	Rows, Cols          uint16
}

// Next TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.SizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Read(p []byte) (int, error) {
	session := TerminalSessions.Get(t.Id)
	if session.TimeOut.Before(time.Now()) {
		_ = TerminalSessions.Sessions[session.Id].sockJSSession.Close(2, "the connection has been disconnected. Please reconnect")
		return 0, errors.New("the connection has been disconnected. Please reconnect")
	}
	TerminalSessions.Set(session.Id, session)
	var reply string
	var msg map[string]uint16
	reply, err := session.sockJSSession.Recv()
	if err != nil {
		// Send terminated signal to process to avoid resource leak
		return copy(p, END_OF_TRANSMISSION), err
	}

	if err := json.Unmarshal([]byte(reply), &msg); err != nil {
		return copy(p, reply), nil
	} else {
		session.SizeChan <- remotecommand.TerminalSize{
			Width:  msg["cols"],
			Height: msg["rows"],
		}
		return 0, nil
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t TerminalSession) Write(p []byte) (int, error) {
	session := TerminalSessions.Get(t.Id)
	if session.TimeOut.Before(time.Now()) {
		_ = TerminalSessions.Sessions[session.Id].sockJSSession.Close(2, "the connection has been disconnected. Please reconnect")
		return 0, errors.New("the connection has been disconnected. Please reconnect")
	}
	TerminalSessions.Set(session.Id, session)
	err := session.sockJSSession.Send(string(p))
	return len(p), err
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t TerminalSession) Toast(p string) error {

	if err := t.sockJSSession.Send(p); err != nil {
		return err
	}
	return nil
}

// SessionMap stores a map of all TerminalSession objects and a lock to avoid concurrent conflict
type SessionMap struct {
	Sessions map[string]TerminalSession
	Lock     sync.RWMutex
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) TerminalSession {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	return sm.Sessions[sessionId]
}

// Set store a TerminalSession to SessionMap
func (sm *SessionMap) Set(sessionId string, session TerminalSession) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	session.TimeOut = time.Now().Add(SessionTerminalStoreTime * time.Minute)
	sm.Sessions[sessionId] = session
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
	if _, ok := sm.Sessions[sessionId]; !ok {
		return
	}
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	session := sm.Sessions[sessionId]
	err := session.sockJSSession.Close(status, reason)
	if err != nil && status != 1 {
		log.Println(err)
	}
	delete(sm.Sessions, sessionId)
}

// Clean all session when system logout
func (sm *SessionMap) Clean() {
	for _, v := range sm.Sessions {
		err := v.sockJSSession.Close(2, "system is logout, please retry...")
		if err != nil {
			return
		}
	}
	sm.Sessions = make(map[string]TerminalSession)
}

var TerminalSessions = SessionMap{Sessions: make(map[string]TerminalSession)}

// handleTerminalSession is Called by net/http for any new /api/sockjs connections
func handleTerminalSession(session sockjs.Session) {
	var (
		buf             string
		err             error
		msg             TerminalMessage
		terminalSession TerminalSession
	)

	if buf, err = session.Recv(); err != nil {
		log.Printf("handleTerminalSession: can't Recv: %v", err)
		return
	}

	if err = json.Unmarshal([]byte(buf), &msg); err != nil {
		log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}

	if msg.Op != "bind" {
		log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		return
	}

	if terminalSession = TerminalSessions.Get(msg.SessionID); terminalSession.Id == "" {
		log.Printf("handleTerminalSession: can't find session '%s'", msg.SessionID)
		return
	}

	terminalSession.sockJSSession = session
	TerminalSessions.Set(msg.SessionID, terminalSession)
	terminalSession.Bound <- nil
}

// CreateAttachHandler is called from main for /api/sockjs
func CreateAttachHandler(path string) http.Handler {
	return sockjs.NewHandler(path, sockjs.DefaultOptions, handleTerminalSession)
}

func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config, cmd []string, namespace string, podName string, containerName string, ptyHandler PtyHandler) error {

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

func start(session TerminalSession, containerName string, cmd string) {
	client := session.RequestInfo.K8sClient
	cfg := session.RequestInfo.Cfg
	namespace := session.RequestInfo.Namespace
	pod := session.RequestInfo.PodName
	var err error
	validShells := []string{cmd}
	if isValidShell(validShells, cmd) {
		err = startProcess(client, cfg, validShells, namespace, pod, containerName, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for _, testShell := range validShells {
			script := []string{testShell}
			if err = startProcess(client, cfg, script, namespace, pod, containerName, session); err == nil {
				break
			}
		}
	}
	if err != nil {
		err := session.sockJSSession.Send(err.Error())
		if err != nil {
			return
		}
	}
}
func GenTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

// isValidShell checks if the shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

func (self TerminalSession) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	sessionId := request.FormValue("sessionId")
	cmd := request.FormValue("cmd")
	container := request.FormValue("container")
	if sessionId == "" {
		log.Printf("handleTerminalSession: sessionId is empty")
		return
	}
	if cmd == "sh" {
		cmd = "/bin/sh"
	} else {
		cmd = "/bin/bash"
	}
	terminalSession := TerminalSessions.Get(sessionId)
	if terminalSession.Id == "" {
		log.Printf("can't find session '%s'", sessionId)
		return
	}
	sessionHandler := func(session sockjs.Session) {
		terminalSession.sockJSSession = session
		TerminalSessions.Set(sessionId, terminalSession)
		start(terminalSession, container, cmd)
	}
	sockjs.NewHandler("/terminal/sockjs", sockjs.DefaultOptions, sessionHandler).ServeHTTP(w, request)
}
