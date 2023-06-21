package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

type TokenEntity struct {
	ApiServer string
	Token     string
	Timestamp int64
	Config    *rest.Config
}

// TokenCache token 缓存
var TokenCache = make(map[string]TokenEntity)

// Timeout token 超时时间
const Timeout = 30 * 60 * 100

func (this TokenEntity) CheckToken() bool {
	currentTime := time.Now().UnixMilli()
	timestamp := this.Timestamp
	if currentTime-timestamp >= Timeout {
		delete(TokenCache, SumMD5(this.Token))
		return false
	}
	return true
}

func RestConfigByToken(apiServer string, token string) (*rest.Config, error) {
	return &rest.Config{
		Host:            apiServer,
		BearerToken:     token,
		BearerTokenFile: "",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // 设置为true时 不需要CA
		},
	}, nil
}

func SumMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckTokenTimeout() {
	fmt.Println("hello", time.Now().UnixMilli())
	currentTime := time.Now().UnixMilli()
	for key, token := range TokenCache {
		timestamp := token.Timestamp
		if currentTime-timestamp >= Timeout {
			delete(TokenCache, key)
		}
	}
}

func GetContainer(k8sClient kubernetes.Interface, namespace string, pod string) (*v1.Pod, error) {
	return k8sClient.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
}
