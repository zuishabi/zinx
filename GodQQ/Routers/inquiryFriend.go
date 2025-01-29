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

type InquiryFriendRouter struct {
	znet.BaseRouter
}

func (i *InquiryFriendRouter) Handle(request ziface.IRequest) {
	fmt.Println("收到请求")
	inquiryMsg := &msg.InquiryFriend{}
	err := proto.Unmarshal(request.GetData(), inquiryMsg)
	if err != nil {
		fmt.Println("unmarshal inquiryFriend err = ", err)
		return
	}
	resultMsg := &msg.ResultFriend{}
	//检查当前用户是否存在
	userInfo := mysqlQQ.UserInfo{}
	err = mysqlQQ.Db.Where("uid = ?", inquiryMsg.FriendId).First(&userInfo).Error
	if err != nil {
		resultMsg.UserId = 0
		core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(13, resultMsg)
		return
	}
	//当前用户存在
	var firstUser, secondUser uint32
	if inquiryMsg.UserId > inquiryMsg.FriendId {
		firstUser = inquiryMsg.UserId
		secondUser = inquiryMsg.FriendId
	} else {
		firstUser = inquiryMsg.FriendId
		secondUser = inquiryMsg.UserId
	}
	//在数据库中查询对应的数据并返回给客户端
	res := mysqlQQ.FriendsList{}
	resultMsg.UserId = inquiryMsg.FriendId
	resultMsg.UserName = userInfo.UserName
	err = mysqlQQ.Db.Where("big_id = ?", firstUser).Where("small_id = ?", secondUser).First(&res).Error
	if err != nil {
		//没有找到对应的数据
		(*resultMsg).IsFriend = false
	} else {
		//找到了对应的数据
		resultMsg.IsFriend = res.IsFriend
	}
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(13, resultMsg)
}
