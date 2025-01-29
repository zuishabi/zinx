package Routers

import (
	"zinx/GodQQ/core"
	"zinx/ziface"
	"zinx/znet"
)

type UserReadyRouter struct {
	znet.BaseRouter
}

func (u *UserReadyRouter) Handle(request ziface.IRequest) {
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	user.OnUserReady()
}
