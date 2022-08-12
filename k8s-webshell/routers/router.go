package routers

import (
	"github.com/astaxie/beego"
	"kube-terminal/k8s-webshell/client/logging"
	"kube-terminal/k8s-webshell/client/terminal"
	"kube-terminal/k8s-webshell/controllers"
)

func init() {
	beego.Router("/", &controllers.HomeController{})
	beego.Router("/terminal", &controllers.TerminalController{}, "get:Get")

	beego.Router("/terminal/pod", &controllers.TerminalController{}, "get:Terminal")

	beego.Router("/pod/terminal/:sessionId/:shell", &controllers.TerminalController{}, "get:TerminalView")

	beego.Router("/terminal/token", &controllers.TokenController{}, "post:Token")
	//POST 传递apiServer、k8sToken、shell 等信息，获取对应 session
	beego.Router("/session/get/:shell", &controllers.SessionController{}, "post:GetSession")

	beego.Handler("/terminal/ws", &controllers.TerminalSockjs{}, true)
	beego.Handler("/logging/sockjs/", logging.LogSession{}, true)
	beego.Handler("/terminal/sockjs/", terminal.TerminalSession{}, true)
}
