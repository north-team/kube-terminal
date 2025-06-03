package logging

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/siddontang/go/log"
	"time"
)

type RedisSessionStore struct {
	redis redis.UniversalClient
}

func NewRedisSessionStore(client redis.UniversalClient) *RedisSessionStore {
	return &RedisSessionStore{redis: client}
}

func (s *RedisSessionStore) Get(id string) *LogSession {
	key := "kube-support:session:" + id
	data, err := s.redis.Get(key).Bytes()
	if err != nil {
		log.Error("获取 " + key + " 失败" + err.Error())
		panic("Failed to get session, " + err.Error())
	}
	var sess *LogSession
	err = json.Unmarshal(data, &sess)
	if err != nil {
		log.Error("获取 " + key + " 失败" + err.Error())
		panic("Failed to get session, " + err.Error())
	}
	return sess
}

func (s *RedisSessionStore) Set(id string, session *LogSession) {
	data, err := json.Marshal(session)
	if err != nil {
		panic("Failed to set session, " + err.Error())
	}
	if err := s.redis.Set("kube-support:session:"+id, data, time.Hour).Err(); err != nil {
		panic("Failed to set session, " + err.Error())
	}
}

func (s *RedisSessionStore) Close(id, reason string, status uint32) {
	// 不支持 sockjs.Session 关闭（这通常是服务内状态）
	// 所以仅删除 Redis 中的记录
	if err := s.redis.Del("kube-support:session:" + id).Err(); err != nil {
		panic("Failed to close session, " + err.Error())
	}
}

func (s *RedisSessionStore) Clean() {
	// 遍历 keys 并删除（注意：生产慎用 KEYS）
	keys, err := s.redis.Keys("kube-support:session:*").Result()
	if err != nil {
		panic("Failed to clean session, " + err.Error())
		return
	}
	for _, k := range keys {
		s.redis.Del(k)
	}
}
