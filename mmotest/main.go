package main

import (
	"zinx/mmotest/core"
	"zinx/ziface"
	"zinx/znet"
)

func OnConnectionAdd(conn ziface.IConnection) {
	//创建一个Player对象
	player := core.NewPlayer(conn)
	//给客户端发送msgID：1的消息:同步当前player的id给客户端
	player.SyncPid()
	//给客户端发送msgID：200的消息,同步当前Player的初始位置给客户端
	player.BroadCastStartPosition()
	//将当前新上线的玩家添加到WorldManager中
	core.WorldMgrObj.AddPlayer(player)
	//将该链接绑定一个Pid玩家ID的属性
	conn.SetProperty("pid", player.Pid)
}

func main() {
	//创建zinx server句柄
	server := znet.NewServer()
	//注册客户端连接和断开触发的函数
	server.SetOnConnStart(OnConnectionAdd)
	//注册服务器路由

	//启动服务器
	server.Serve()
}
