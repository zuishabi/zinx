package main

import (
	"fmt"
	"zinx/GodQQ/Routers"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
	"zinx/ziface"
	"zinx/znet"
)

// 当连接建立时调用的函数
func OnConnStart(conn ziface.IConnection) {
}

// 当连接丢失时调用的函数
func OnConnStop(conn ziface.IConnection) {
	user := core.IOnlineMap.GetUserByConn(conn)
	if user != nil {
		core.IOnlineMap.RemoveUser(user.Uid)
		onOrOffLine := &msg.OnOrOffLineMsg{
			Uid:      user.Uid,
			UserName: user.UserName,
			Type:     false,
		}
		core.IOnlineMap.BroadCast(5, onOrOffLine)
	}
}

func main() {
	defer core.MainRedisConn.Close() //当服务器关闭时关闭redis的主连接
	defer redisQQ.Pool.Close()       //当服务器关闭时停止连接池
	//检测是否连接至redis服务器
	_, err := core.MainRedisConn.Do("Ping")
	if err != nil {
		fmt.Println("connect to redis err = ", err)
		return
	} else {
		fmt.Println("success connect to redis")
	}
	//开启服务器
	server := znet.NewServer()
	server.SetOnConnStart(OnConnStart)
	server.SetOnConnStop(OnConnStop)
	server.AddRouter(0, &Routers.LoginRouter{})
	server.AddRouter(1, &Routers.BroadCastRouter{})
	server.AddRouter(2, &Routers.RegisterRouter{})
	server.AddRouter(3, &Routers.PrivateChatRouter{})
	server.AddRouter(4, &Routers.SendOnlineUsersRouter{})
	server.AddRouter(6, &Routers.GenerateCaptchaRouter{})
	server.Serve()
}
