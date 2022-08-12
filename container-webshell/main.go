package main

import (
	"github.com/astaxie/beego"
	_ "kube-terminalcontainer-webshell/routers"
)

func main() {
	beego.Run()
}
