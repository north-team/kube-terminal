package routers

import (
	"github.com/astaxie/beego"
	"kube-terminalcontainer-webshell/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/ws", &controllers.Wscontroller{})
}
