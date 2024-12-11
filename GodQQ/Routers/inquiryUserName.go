package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type InquiryUserNameRouter struct {
	znet.BaseRouter
}

func (i *InquiryUserNameRouter) Handle(request ziface.IRequest) {
	req := msg.InquiryUser{}
	proto.Unmarshal(request.GetData(), &req)
	inquiryUser := msg.InquiryUser{}
	targetUser := mysqlQQ.UserInfo{}
	mysqlQQ.Db.Where("uid = ?", req.GetUserId()).First(&targetUser)
	if inquiryUser.GetUserId() == 0 {
		fmt.Println("Inquiry user error")
		return
	}
	inquiryUser.UserName = targetUser.UserName
	inquiryUser.UserId = targetUser.UID
	fmt.Println(inquiryUser.UserName)
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(200, &inquiryUser)
}
