package core

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

// 管理全局

var MainRedisConn redis.Conn

// 获得一个可用的uid
func GetUid() uint32 {
	reply, err := MainRedisConn.Do("get", "current_uid")
	if err != nil {
		fmt.Println("Get Uid from redis err = ", err)
		return 0
	}
	str, err := redis.String(reply, err)
	if err != nil {
		fmt.Println("[GetUid] reply to string err = ", err)
		return 0
	}
	uid64, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		fmt.Println("[GetUid] string to uint64 err = ", err)
		return 0
	}
	uid32 := uint32(uid64) + 1
	_, err = MainRedisConn.Do("set", "current_uid", uid32)
	if err != nil {
		fmt.Println("[GetUid] set current_uid err = ", err)
		return 0
	}
	return uid32
}
