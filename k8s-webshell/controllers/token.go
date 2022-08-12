package controllers

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kube-terminal/k8s-webshell/client"
	"time"
)

type TokenController struct {
	BaseController
}

// Token API请求获得Token
func (self *TokenController) Token() {
	currentTime := time.Now().UnixMilli()
	apiServer := self.GetString("apiServer")
	k8sToken := self.GetString("k8sToken")
	if apiServer == "" || k8sToken == "" {
		self.SimpleErrorJson(500, "apiServer 和 k8sToken 不能为空")
	}
	token := client.SumMD5(k8sToken)
	cacheToken, ok := client.TokenCache[token]
	if ok {
		timestamp := cacheToken.Timestamp
		if currentTime-timestamp < client.Timeout {
			self.SuccessJson(token)
		} else {
			delete(client.TokenCache, token)
		}
	}
	//验证apiServer 和 k8sToken 是否有效
	config, err := client.RestConfigByToken(apiServer, k8sToken)
	if err != nil {
		self.ErrorJson(500, "获取k8s客户端失败", err)
	}
	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		self.ErrorJson(500, "获取k8s客户端失败", err)
	}
	_, err = restClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		self.ErrorJson(500, "K8S账户校验失败", err)
	}
	client.TokenCache[token] = client.TokenEntity{
		ApiServer: apiServer,
		Token:     k8sToken,
		Timestamp: currentTime,
		Config:    config,
	}
	self.SuccessJson(token)
}
