package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/spf13/cast"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
	"kube-terminal/client"
	"net/http"
	"reflect"
)

type TerminalSockjs struct {
	conn      sockjs.Session
	sizeChan  chan *remotecommand.TerminalSize
	namespace string
	pod       string
	container string
	apiServer string
	token     string
	cmd       string
	shell     string
	tailLines *int64
	follow    bool
}

func (self TerminalSockjs) Read(p []byte) (int, error) {
	var reply string
	var msg map[string]uint16
	reply, err := self.conn.Recv()
	if err != nil {
		return 0, err
	}
	if err := json.Unmarshal([]byte(reply), &msg); err != nil {
		return copy(p, reply), nil
	} else {
		self.sizeChan <- &remotecommand.TerminalSize{
			Width:  msg["cols"],
			Height: msg["rows"],
		}
		return 0, nil
	}
}

func (self TerminalSockjs) Write(p []byte) (int, error) {
	err := self.conn.Send(string(p))
	return len(p), err
}

// Next 实现tty size queue
func (self *TerminalSockjs) Next() *remotecommand.TerminalSize {
	size := <-self.sizeChan
	beego.Debug(fmt.Sprintf("terminal size to width: %d height: %d", size.Width, size.Height))
	return size
}

func restClientByToken(env *TerminalSockjs) (*rest.Config, error) {
	return &rest.Config{
		Host:            env.apiServer,
		BearerToken:     env.token,
		BearerTokenFile: "",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // 设置为true时 不需要CA
		},
	}, nil
}

// kubeConfig方式认证
func buildConfigFromContextFlags(context, kubeConfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

// Handler 处理输入输出与sockjs 交互
func Handler(t *TerminalSockjs) error {
	config, _ := restClientByToken(t)
	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	req := buildPodWSRequest(t, restClient)
	executor, err := remotecommand.NewSPDYExecutor(
		config, http.MethodPost, req.URL(),
	)
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             t,
		Stdout:            t,
		Stderr:            t,
		Tty:               true,
		TerminalSizeQueue: t,
	})
}

func buildPodWSRequest(t *TerminalSockjs, restClient *kubernetes.Clientset) *rest.Request {
	if t.shell == "log" {
		logOpt := &v1.PodLogOptions{
			Container: t.container,
			Follow:    t.follow,
			TailLines: t.tailLines,
			Previous:  false,
		}
		req := restClient.CoreV1().RESTClient().Get().
			Resource("pods").
			SubResource("log").
			Name(t.pod).
			Namespace(t.namespace).
			VersionedParams(
				logOpt,
				scheme.ParameterCodec,
			)
		return req
	}
	req := restClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(t.pod).
		Namespace(t.namespace).
		SubResource("exec").
		Param("container", t.container).
		Param("stdout", "true").
		Param("stderr", "true").
		Param("stdin", "true").
		Param("command", t.cmd).
		Param("tty", "true")
	req.VersionedParams(
		&v1.PodExecOptions{
			Container: t.container,
			Command:   []string{},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		},
		scheme.ParameterCodec,
	)
	return req
}

// 实现http.handler 接口获取入参
func (self TerminalSockjs) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	namespace := request.FormValue("namespace")
	pod := request.FormValue("pod")
	container := request.FormValue("container")
	token := request.FormValue("token")
	shell := request.FormValue("shell")
	follow := request.FormValue("follow")
	tailLines := cast.ToInt64(request.FormValue("tailLines"))

	if token == "" {
		token = request.Header.Get("token")
	}
	cmd := request.FormValue("cmd")
	if cmd == "" || cmd == "sh" {
		cmd = "/bin/sh"
	} else {
		cmd = "/bin/bash"
	}
	tokenEntity := client.TokenCache[token]
	if reflect.DeepEqual(tokenEntity, client.TokenEntity{}) {
		beego.Error("TOKEN 不存在！")
	}
	apiServer := tokenEntity.ApiServer
	k8sToken := tokenEntity.Token
	sockjsHandler := func(session sockjs.Session) {
		t := &TerminalSockjs{session, make(chan *remotecommand.TerminalSize),
			namespace, pod, container, apiServer, k8sToken, cmd, shell, &tailLines, cast.ToBool(follow)}
		if err := Handler(t); err != nil {
			beego.Error(err)
		}
	}

	sockjs.NewHandler("/terminal/ws", sockjs.DefaultOptions, sockjsHandler).ServeHTTP(w, request)
}
