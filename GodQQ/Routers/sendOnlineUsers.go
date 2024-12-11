package Routers

import (
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type SendOnlineUsersRouter struct {
	znet.BaseRouter
}

func (s *SendOnlineUsersRouter) Handle(request ziface.IRequest) {
	onlineUsersMsg := &msg.OnlineUsers{}
	uids := make([]uint32, 0)
	names := make([]string, 0)
	for uid, user := range core.IOnlineMap.UserMap {
		uids = append(uids, uid)
		names = append(names, user.UserName)
	}
	onlineUsersMsg.Uid = uids
	onlineUsersMsg.UserName = names
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(4, onlineUsersMsg)
}
