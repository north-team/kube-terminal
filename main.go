package main

import (
	"github.com/astaxie/beego"
	"github.com/siddontang/go/log"
	_ "kube-terminal/routers"
	"kube-terminal/utils/config"
	"kube-terminal/utils/redis"
)

func main() {
	redisConfig := config.LoadConfig("/opt/fit2cloud/conf/fit2cloud.properties")
	// 初始化 Redis
	if err := redis.InitRedis(redisConfig); err != nil {
		panic("Redis 初始化失败: " + err.Error())
	}

	if redis.RedisClient == nil {
		log.Error("配置文件 /opt/fit2cloud/conf/fit2cloud.properties，没有指定 Redis Mode 不使用 Redis 存储 session")
	}
	// 初始化 session 存储
	redis.InitSessionStore()

	beego.SetStaticPath("/kube-terminal/static", "static")
	beego.Run()
}
