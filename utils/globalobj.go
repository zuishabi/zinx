package utils

import (
	"encoding/json"
	"github.com/zuishabi/zinx/ziface"
	"os"
	"runtime"
)

// 存储一切有关zinx框架的全局参数，供其他模块使用，一些参数是可以通过zinx.json由用户进行配置
type GlobalObj struct {
	//server
	TcpServer        ziface.IServer //当前zinx全局的server对象
	Host             string         //当前服务器主机监听的地址
	TcpPort          int            //当前服务器主机监听的端口
	Name             string         //当前服务器的名称
	WorkerPoolSize   uint32         //当前业务工作Worker池的goroutine的数量
	MaxWorkerTaskLen uint32         //每个worker对应的消息队列的任务的数量的最大值
	//zinx
	Version        string //zinx的版本号
	MaxConn        int    //当前服务器主机允许的最大连接数
	MaxPackageSize uint32 //当前zinx框架数据包的最大值
	MaxMsgChanLen  uint32 //SendBuffMsg发送消息的缓冲最大长度
}

// 定义一个全局的对外GlobalObj
var GlobalObject *GlobalObj

// 从zinx.json去加载用户自定义参数
func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}
	//将json文件中的数据解析到struct中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

// 提供一个init方法，初始化当前的GlobalObj
func init() {
	//如果配置文件没有加载，默认的值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServer",
		Version:          "v0.4",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPackageSize:   4096,
		WorkerPoolSize:   uint32(runtime.NumCPU()),
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
	}
	//应该尝试从conf/zinx.json中去加载一些用户自定义的参数
	GlobalObject.Reload()
}
