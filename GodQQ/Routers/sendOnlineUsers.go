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
	userNames := make([]string, 0)
	uids := make([]uint32, 0)
	for uid, user := range core.IOnlineMap.UserMap {
		userNames = append(userNames, user.UserName)
		uids = append(uids, uid)
	}
	onlineUsersMsg.Uid = uids
	onlineUsersMsg.Name = userNames
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(4, onlineUsersMsg)
}
