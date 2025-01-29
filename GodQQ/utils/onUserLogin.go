package utils

import (
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
)

//当连接建立时调用的所有函数

func init() {
	core.FunctionLists = append(core.FunctionLists, SendFriendList)
	core.FunctionLists = append(core.FunctionLists, SendAddFriendInfo)
}

// 向客户端发送全部的好友请求信息
func SendAddFriendInfo(user *core.User) {
	addFriendMsg := msg.AddFriend{}
	AddFriendRequests := make([]mysqlQQ.AddFriendList, 0)
	mysqlQQ.Db.Where("target_id = ?", user.Uid).Find(&AddFriendRequests)
	for _, i := range AddFriendRequests {
		addFriendMsg.Info = i.Info
		addFriendMsg.SourceId = i.SourceID
		addFriendMsg.TargetId = i.TargetID
		addFriendMsg.Type = true
		user.SendMsg(14, &addFriendMsg)
	}
}

// 向客户端发送好友列表
func SendFriendList(user *core.User) {
	friendsList1 := make([]mysqlQQ.FriendsList, 0)
	friendsList2 := make([]mysqlQQ.FriendsList, 0)
	mysqlQQ.Db.Where("small_id = ?", user.Uid).Where("big_id > ?", user.Uid).Where("is_friend = ?", true).Find(&friendsList1)
	mysqlQQ.Db.Where("big_id = ?", user.Uid).Where("small_id < ?", user.Uid).Where("is_friend = ?", true).Find(&friendsList2)
	ids := make([]uint32, 0)
	for _, i := range friendsList1 {
		ids = append(ids, i.BigID)
	}
	for _, i := range friendsList2 {
		ids = append(ids, i.SmallID)
	}
	friendsList := msg.GetFriendsList{
		UserIds: ids,
	}
	user.SendMsg(15, &friendsList)
}
