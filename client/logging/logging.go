package logging

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"kube-terminal/model/request"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func GenLoggingSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

type LogSession struct {
	Id            string
	Bound         chan error
	sockJSSession sockjs.Session
	RequestInfo   request.TerminalRequest
}

type SessionMap struct {
	Sessions map[string]LogSession
	Lock     sync.Mutex
}

func (sm *SessionMap) Get(sessionId string) LogSession {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	return sm.Sessions[sessionId]
}

func (sm *SessionMap) Set(sessionId string, session LogSession) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Sessions[sessionId] = session
}

func (sm *SessionMap) Close(sessionId, reason string, status uint32) {
	if _, ok := sm.Sessions[sessionId]; !ok {
		return
	}
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	err := sm.Sessions[sessionId].sockJSSession.Close(status, reason)
	if err != nil {
		log.Println(err)
	}
	delete(sm.Sessions, sessionId)
}

func (sm *SessionMap) Clean() {
	for _, v := range sm.Sessions {
		v.sockJSSession.Close(2, "system is logout, please retry...")
	}
	sm.Sessions = make(map[string]LogSession)
}

var LogSessions = SessionMap{Sessions: make(map[string]LogSession)}

type LogMessage struct {
	SessionID string
	Data      string
}

func CreateLoggingHandler(path string) http.Handler {
	return sockjs.NewHandler(path, sockjs.DefaultOptions, logHandler)
}

func logHandler(session sockjs.Session) {

	var (
		buf        string
		err        error
		msg        LogMessage
		logSession LogSession
	)
	if buf, err = session.Recv(); err != nil {
		log.Printf("handleLogSession: can't Recv: %v", err)
		return
	}
	if err = json.Unmarshal([]byte(buf), &msg); err != nil {
		log.Printf("handleLogSession: can't UnMarshal (%v): %s", err, buf)
		return
	}
	if logSession = LogSessions.Get(msg.SessionID); logSession.Id == "" {
		log.Printf("handleLogSession: can't find session '%s'", msg.SessionID)
		return
	}
	logSession.sockJSSession = session
	LogSessions.Set(msg.SessionID, logSession)
	logSession.Bound <- nil
}

func WaitForLoggingStream(k8sClient kubernetes.Interface, namespace string, pod string, container string, tailLines int64, follow bool, sessionId string) {
	select {
	case <-LogSessions.Get(sessionId).Bound:
		//close(LogSessions.Get(sessionId).Bound)
		err := startLogProcess(k8sClient, namespace, pod, container, tailLines, follow, LogSessions.Get(sessionId))
		if err != nil {
			LogSessions.Close(sessionId, err.Error(), 2)
			return
		}
		LogSessions.Close(sessionId, "Process exited", 1)
	}
}

func WaitForTerminalBySession(session LogSession, containerName string, tailLines int64, follow bool) {
	client := session.RequestInfo.K8sClient
	namespace := session.RequestInfo.Namespace
	pod := session.RequestInfo.PodName
	id := session.Id
	WaitForLoggingStream(client, namespace, pod, containerName, tailLines, follow, id)
}

func startLogProcess(k8sClient kubernetes.Interface, namespace string, pod string, container string, tailLines int64, follow bool, session LogSession) error {
	fmt.Println("tailLines======", tailLines)
	reader, err := k8sClient.CoreV1().
		Pods(namespace).
		GetLogs(pod, &v1.PodLogOptions{
			Container: container,
			Follow:    follow,
			TailLines: &tailLines,
		}).Stream(context.TODO())
	if err != nil {
		return err
	}

	ss := session.sockJSSession
	for {
		buf := make([]byte, 2048)
		numBytes, err := reader.Read(buf)
		if numBytes > 0 {
			message := string(buf[:numBytes])
			for _, val := range strings.Split(message, "\n") {
				if strings.TrimSpace(val) == "" {
					continue
				}
				err = ss.Send(strings.TrimSpace(val) + "\r\n")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		if err != nil {
			return err
		}
	}
}

func start(session LogSession, containerName string, tailLines int64, follow bool) {
	client := session.RequestInfo.K8sClient
	namespace := session.RequestInfo.Namespace
	pod := session.RequestInfo.PodName
	err := startLogProcess(client, namespace, pod, containerName, tailLines, follow, session)
	if err != nil {
		err := session.sockJSSession.Send(err.Error())
		if err != nil {
			return
		}
	}
}

func (self LogSession) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	sessionId := request.FormValue("sessionId")
	container := request.FormValue("container")
	tailLines := request.FormValue("tailLines")
	follow := request.FormValue("follow")
	if sessionId == "" {
		log.Printf("handleTerminalSession: sessionId is empty")
		return
	}
	logSession := LogSessions.Get(sessionId)
	if logSession.Id == "" {
		log.Printf("can't find session '%s'", sessionId)
		return
	}
	sessionHandler := func(session sockjs.Session) {
		logSession.sockJSSession = session
		LogSessions.Set(sessionId, logSession)
		start(logSession, container, cast.ToInt64(tailLines), cast.ToBool(follow))
	}
	sockjs.NewHandler("/logging/sockjs", sockjs.DefaultOptions, sessionHandler).ServeHTTP(w, request)
}
