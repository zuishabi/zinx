package redisQQ

import (
	"github.com/gomodule/redigo/redis"
)

var Pool *redis.Pool

func init() {
	Pool = &redis.Pool{
		MaxIdle:     8,   //最大空闲连接数
		MaxActive:   0,   //表示和数据库的最大链接数，0表示不限制
		IdleTimeout: 300, //最大空闲时间
		Dial: func() (redis.Conn, error) { //初始化连接的函数，连接哪个ip的redis
			return redis.Dial("tcp", "127.0.0.1:6379", redis.DialPassword("861214959"))
		},
	}
}
