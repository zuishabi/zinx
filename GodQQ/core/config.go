package core

import (
	"encoding/json"
	"fmt"
	"os"
	"zinx/GodQQ/redisQQ"
)

var ConfigObj *Config

type Config struct {
	IsRefresh bool `json:"is_refresh"` //判断服务器是否需要初始化，即更新数据库中的内容
	IsUpdate  bool `json:"is_updated"` //判断服务器是否需要更新，即更新数据库等操作，且不用删除原先的数据
}

func init() {
	MainRedisConn = redisQQ.Pool.Get()
	ConfigObj = &Config{}
	load()
	if ConfigObj.IsRefresh {
		MainRedisConn.Do("flushdb")
		MainRedisConn.Do("set", "current_uid", 0)
		fmt.Println("[GodQQ Redis] : has refresh...")
	}
}

func load() {
	data, err := os.ReadFile("conf/GodQQ.json")
	if err != nil {
		panic(err)
	}
	//将json文件中的数据解析到struct中
	err = json.Unmarshal(data, ConfigObj)
	if err != nil {
		panic(err)
	}
}
