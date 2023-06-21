package main

import (
	"github.com/astaxie/beego"
	_ "kube-terminal/routers"
)

func main() {
	beego.Run()
}
