package controllers

import (
	"context"
	"github.com/siddontang/go/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kube-terminal/client"
	"kube-terminal/client/logging"
	"kube-terminal/client/terminal"
	"kube-terminal/model/request"
	"reflect"
)

type TerminalController struct {
	BaseController
}

func (self *TerminalController) Get() {
	self.Data["apiServer"] = self.GetString("apiServer")
	self.Data["token"] = self.GetString("token")
	self.Data["namespace"] = self.GetString("namespace")
	self.Data["pod"] = self.GetString("pod")
	self.Data["container"] = self.GetString("container")
	self.TplName = "terminal.html"
}

func (self *TerminalController) Terminal() {
	token := self.GetString("token")
	namespace := self.GetString("namespace")
	pod := self.GetString("pod")

	self.Data["token"] = token
	self.Data["namespace"] = namespace
	self.Data["pod"] = pod
	self.Data["shell"] = self.GetString("shell")
	self.Data["container"] = self.GetString("container")

	//err, podDetail := self.getPodDetail(token, namespace, pod)
	//if err != nil {
	//	self.ErrorJson(500, "", nil)
	//}
	//self.Data["podDetail"] = podDetail
	self.TplName = "index.html"
}

type ResponseData struct {
	SessionId string
	Shell     string
	PodDetail *v1.Pod
}

func (self *TerminalController) TerminalView() {
	sessionId := self.GetString(":sessionId")
	shell := self.GetString(":shell")
	if shell == "" || sessionId == "" {
		self.ErrorJson(401, "sessionId 和 shell 不能为空！", nil)
	}
	var info request.TerminalRequest
	if shell == "log" {
		logSession := logging.LogSessions.Get(sessionId)
		info = logSession.RequestInfo
		if reflect.DeepEqual(info, request.TerminalRequest{}) {
			self.ErrorJson(401, "未认证！", nil)
		}
	} else {
		terminalSession := terminal.TerminalSessions.Get(sessionId)
		info = terminalSession.RequestInfo
		if reflect.DeepEqual(info, request.TerminalRequest{}) {
			self.ErrorJson(401, "未认证！", nil)
		}
	}

	self.Data["shell"] = shell
	self.Data["sessionId"] = sessionId
	podDetail, err := client.GetContainer(info.K8sClient, info.Namespace, info.PodName)
	if err != nil {
		self.ErrorJson(500, "无法获取POD中的容器，连接失败！", nil)
	}
	self.Data["podDetail"] = podDetail
	self.TplName = "pod-terminal.html"
}

func (self *TerminalController) getPodDetail(token string, namespace string, pod string) (error, *v1.Pod) {
	k8sToken := client.TokenCache[token]
	if reflect.DeepEqual(k8sToken, client.TokenEntity{}) {
		log.Error("TOKEN 不存在！")
		self.SimpleErrorJson(500, "TOKEN 不存在！")
	}
	if !k8sToken.CheckToken() {
		log.Error("TOKEN 已过期！")
		self.SimpleErrorJson(500, "TOKEN 已过期！")
	}
	restClient, err := kubernetes.NewForConfig(k8sToken.Config)
	if err != nil {
		self.ErrorJson(500, "", nil)
	}
	podDetail, err := restClient.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	return err, podDetail
}
