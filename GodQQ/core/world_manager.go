package core

import (
	"github.com/gomodule/redigo/redis"
)

// 管理全局

var MainRedisConn redis.Conn
