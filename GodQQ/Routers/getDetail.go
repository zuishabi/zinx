package Routers

import (
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type GetDetailRouter struct {
	znet.BaseRouter
}

func (g *GetDetailRouter) Handle(request ziface.IRequest) {
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	getShareDetail := msg.GetShareDetail{}
	proto.Unmarshal(request.GetData(), &getShareDetail)
	id := getShareDetail.GetId()
	shareInfo := mysqlQQ.ShareInfo{}
	result := mysqlQQ.Db.Where("id = ?", id).Find(&shareInfo)
	sendShareDetail := msg.SendShareDetail{}
	//如果寻找有误，则向客户端返回查询失败
	if result.Error != nil {
		sendShareDetail.Exist = false
		user.SendMsg(10, &sendShareDetail)
		return
	}
	//当更新的时间和客户端更新的时间不同，则服务器上的数据发生了更新，需要重新发送
	if shareInfo.UpdatedAt.Format("2006.01.02.15.04.05") != getShareDetail.UpdatedTime {
		sendShareDetail.Content = shareInfo.Content
	}
	//当客户端向服务器的请求有被隐藏了，则需要重新发送完整内容
	if getShareDetail.IsHidden {
		sendShareDetail.Content = shareInfo.Content
	}
	user.SendMsg(10, &sendShareDetail)
}
