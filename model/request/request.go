package request

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kube-terminal/client"
)

type TerminalRequest struct {
	ApiServer   string               `json:"api_server"`
	Namespace   string               `json:"namespace"`
	PodName     string               `json:"pod_name"`
	k8sClient   kubernetes.Interface `json:"-"`
	Host        string               `json:"host"`
	BearerToken string               `json:"bearer_token"`
	KubeConfig  string               `json:"kube_config"`
}

func (t TerminalRequest) GetConfig() *rest.Config {
	restConfig, err := t.GetRestConfig(t.KubeConfig, t.ApiServer, t.BearerToken)
	if err != nil {
		panic("无法通过 K8S Token 获取有效配置，请检查Token是否正确")
	}
	return restConfig
}

func (t TerminalRequest) GetK8sClient() kubernetes.Interface {
	if t.k8sClient == nil {
		restClient, err := kubernetes.NewForConfig(t.GetConfig())
		t.k8sClient = restClient
		if err != nil {
			panic("无法通过 K8S Token 获取有效配置，请检查Token是否正确")
		}
	}
	return t.k8sClient
}

func (t TerminalRequest) GetRestConfig(kubeConfig string, apiServer string, k8sToken string) (*rest.Config, error) {
	var config *rest.Config
	if kubeConfig != "" {
		// 将kubeconfig字符串内容转换为字节切片
		kubeConfigBytes := []byte(kubeConfig)
		// 使用NewClientConfigFromBytes从kubeconfig字节内容创建ClientConfig实例
		configLoader, err := clientcmd.NewClientConfigFromBytes(kubeConfigBytes)
		if err != nil {
			return nil, err
		}
		// 加载实际的rest.Config
		restConfig, err := configLoader.ClientConfig()
		if err != nil {
			return nil, err
		}
		config = restConfig
	} else {
		//验证apiServer 和 k8sToken 是否有效
		restConfig, err := client.RestConfigByToken(apiServer, k8sToken)
		if err != nil {
			return nil, err
		}
		config = restConfig
	}
	return config, nil
}
