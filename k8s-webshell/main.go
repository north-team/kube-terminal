package main

import (
	"github.com/astaxie/beego"
	_ "kube-terminal/k8s-webshell/routers"
)

func main() {
	beego.Run()
}
