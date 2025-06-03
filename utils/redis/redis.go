package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/siddontang/go/log"
	"kube-terminal/client/logging"
	"kube-terminal/utils/config"
	"strconv"
)

var RedisClient redis.UniversalClient

func InitRedis(conf *config.RedisConfig) error {
	switch conf.Mode {
	case "single":
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     conf.Addr + ":" + strconv.Itoa(conf.Port),
			Password: conf.Password,
			DB:       conf.DB,
		})
	case "sentinel":
		if conf.MasterName == "" || len(conf.Addrs) == 0 {
			return errors.New("哨兵模式必须配置 redis.master 和 redis.addrs")
		}
		RedisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    conf.MasterName,
			SentinelAddrs: conf.Addrs,
			Password:      conf.Password,
			DB:            conf.DB,
		})
	case "cluster":
		if len(conf.Addrs) == 0 {
			return errors.New("集群模式必须配置 redis.addrs")
		}
		RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    conf.Addrs,
			Password: conf.Password,
		})
	default:
		return nil
	}
	if _, err := RedisClient.Ping().Result(); err != nil {
		return err
	}
	return nil
}

func InitSessionStore() {
	if RedisClient != nil {
		logging.LogSessions = logging.NewRedisSessionStore(RedisClient)
		log.Info("Using Redis session store")
	} else {
		logging.LogSessions = logging.NewMemorySessionStore()
		log.Info("Using in-memory session store")
	}
}
