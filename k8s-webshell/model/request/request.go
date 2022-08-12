package request

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type TerminalRequest struct {
	ApiServer string
	Namespace string
	PodName   string
	K8sClient kubernetes.Interface
	Cfg       *rest.Config
}
