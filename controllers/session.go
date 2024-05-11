package controllers

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"kube-terminal/client"
	"kube-terminal/client/logging"
	"kube-terminal/client/terminal"
	"kube-terminal/model/request"
)

type SessionController struct {
	BaseController
}

type TerminalResponse struct {
	ID string `json:"id"`
}

// kubeConfig方式认证
func buildConfigFromContextFlags(context, kubeConfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func (this SessionController) GetSession() {
	apiServer := this.GetString("apiServer")
	k8sToken := this.GetString("k8sToken")
	kubeConfig := this.GetString("kubeConfig")
	namespace := this.GetString("namespace")
	podName := this.GetString("pod")
	shell := this.Ctx.Input.Param(":shell")

	//验证apiServer 和 k8sToken 是否有效
	config, err := client.RestConfigByToken(apiServer, k8sToken)
	// 如果kubeConfig不为空 则获取kubeConfig配置
	if kubeConfig != "" {
		config, err = buildConfigFromContextFlags(kubeConfig, "")
		if err != nil {
			this.ErrorJson(500, "无法通过 K8S Token 获取有效配置，请检查Token是否正确", err)
		}
	}
	if err != nil {
		this.ErrorJson(500, "无法通过 K8S Token 获取有效配置，请检查Token是否正确", err)
	}
	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		this.ErrorJson(500, "无法通过 K8S Token 获取有效配置，请检查Token是否正确", err)
	}
	this.chooseSession(shell, apiServer, namespace, podName, restClient, config)

}

func (this SessionController) chooseSession(shell string, apiServer string, namespace string, podName string, restClient *kubernetes.Clientset, config *rest.Config) {
	if shell == "" {
		this.ErrorJson(500, "shell 值不能为空", nil)
	}
	if shell == "log" {
		sessionId, err := logging.GenLoggingSessionId()
		if err != nil {
			this.ErrorJson(500, "获取 TerminalSession 失败", err)
		}
		logging.LogSessions.Set(sessionId, logging.LogSession{
			Id:    sessionId,
			Bound: make(chan error),
			RequestInfo: request.TerminalRequest{
				ApiServer: apiServer,
				Namespace: namespace,
				PodName:   podName,
				K8sClient: restClient,
				Cfg:       config,
			},
		})
		this.SuccessJson(TerminalResponse{ID: sessionId})
	} else {
		shell = "exec"
		sessionId, err := terminal.GenTerminalSessionId()
		if err != nil {
			this.ErrorJson(500, "获取 TerminalSession 失败", err)
		}
		terminal.TerminalSessions.Set(sessionId, terminal.TerminalSession{
			Id:       sessionId,
			Bound:    make(chan error),
			SizeChan: make(chan remotecommand.TerminalSize),
			RequestInfo: request.TerminalRequest{
				ApiServer: apiServer,
				Namespace: namespace,
				PodName:   podName,
				K8sClient: restClient,
				Cfg:       config,
			},
		})
		this.SuccessJson(TerminalResponse{ID: sessionId})
	}
}
