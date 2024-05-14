package routers

import (
	"github.com/astaxie/beego"
	"kube-terminal/client/logging"
	"kube-terminal/client/terminal"
	"kube-terminal/controllers"
)

func init() {

	ns := beego.NewNamespace("kube-terminal",
		beego.NSRouter("/", &controllers.HomeController{}),
		beego.NSRouter("/terminal", &controllers.TerminalController{}, "get:Get"),
		beego.NSRouter("/terminal/pod", &controllers.TerminalController{}, "get:Terminal"),
		beego.NSRouter("/pod/terminal/:sessionId/:shell", &controllers.TerminalController{}, "get:TerminalView"),
		beego.NSRouter("/terminal/token", &controllers.TokenController{}, "post:Token"),
		beego.NSRouter("/session/get/:shell", &controllers.SessionController{}, "post:GetSession"))
	beego.AddNamespace(ns)
	beego.Handler("/kube-terminal/terminal/ws", &controllers.TerminalSockjs{}, true)
	beego.Handler("/kube-terminal/logging/sockjs/", logging.LogSession{}, true)
	beego.Handler("/kube-terminal/terminal/sockjs/", terminal.TerminalSession{}, true)
}
