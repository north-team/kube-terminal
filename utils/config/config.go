package config

import (
	"github.com/magiconair/properties"
	"kube-terminal/utils/aes"
	"strings"
)

type RedisConfig struct {
	Mode       string // standalone, sentinel, cluster
	Addr       string
	Addrs      []string
	Password   string
	MasterName string
	DB         int
	Port       int
}

func LoadConfig(path string) *RedisConfig {
	p := properties.MustLoadFile(path, properties.UTF8)

	security := p.GetBool("security.password", false)
	mode := p.GetString("redis.mode", "None")
	addr := p.GetString("redis.hostname", "")
	port := p.GetInt("redis.port", 6379)
	redisPassword := p.GetString("redis.password", "")
	var password string
	if security {
		encryptPassword, err := aes.Encrypt(aes.SECRET_PASS, redisPassword)
		if err != nil {
			panic("Redis 解密密码失败: " + err.Error())
		}
		password = encryptPassword
	} else {
		password = redisPassword
	}
	db := p.GetInt("redis.database", 0)
	// 手动解析 addrs 字符串（以逗号分隔）
	addrsRaw := p.GetString("redis.cluster.nodes", "")
	addrs := []string{}
	if addrsRaw != "" {
		addrs = strings.Split(addrsRaw, ",")
	}
	// 手动解析 addrs 字符串（以逗号分隔）
	addrsRaw = p.GetString("redis.sentinel.nodes", "")
	if addrsRaw != "" {
		addrs = strings.Split(addrsRaw, ",")
	}
	master := p.GetString("redis.sentinel.master", "")

	return &RedisConfig{
		Mode:       mode,
		Addr:       addr,
		Password:   password,
		Addrs:      addrs,
		Port:       port,
		MasterName: master,
		DB:         db,
	}
}
