package Routers

import (
	"errors"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"time"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/utils"
	"zinx/ziface"
	"zinx/znet"
)

type AddFriendsRouter struct {
	znet.BaseRouter
}

// 处理添加好友
func (a *AddFriendsRouter) Handle(request ziface.IRequest) {
	addFriend := msg.AddFriend{}
	_ = proto.Unmarshal(request.GetData(), &addFriend)
	addFriendInfo := mysqlQQ.AddFriendList{}
	addFriendInfo.SourceID = addFriend.SourceId
	addFriendInfo.TargetID = addFriend.TargetId
	addFriendInfo.Info = addFriend.GetInfo()
	if addFriend.GetType() {
		//如果为添加好友请求
		//检查是否已有数据
		ifExist := mysqlQQ.AddFriendList{}
		err := mysqlQQ.Db.Where("source_id = ?", addFriend.SourceId).Where("target_id = ?", addFriend.TargetId).First(&ifExist).Error
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			mysqlQQ.Db.Create(&addFriendInfo)
			//检查当前用户是否在线
			if user := core.IOnlineMap.GetUser(addFriend.GetTargetId()); user != nil {
				//当前用户在线
				user.SendMsg(14, &addFriend)
			}
		}
	} else {
		//如果为回应好友请求
		if addFriend.GetRespond() == true {
			//同意好友请求
			info := msg.MessageToClient{
				Uid:       addFriendInfo.TargetID,
				TargetUid: addFriendInfo.SourceID,
				Msg:       &msg.MessageToClient_Text{Text: addFriend.Info},
				Time:      time.Now().Format("2006.01.02.15.04.05"),
				MsgType:   1,
			}
			//发送请求方的打招呼消息
			utils.SendMsgByUser(addFriendInfo.SourceID, addFriendInfo.TargetID, &info)
			//将好友信息添加到数据库
			friendList := mysqlQQ.FriendsList{}
			var big, small uint32
			if addFriendInfo.SourceID > addFriendInfo.TargetID {
				big = addFriendInfo.SourceID
				small = addFriendInfo.TargetID
			} else {
				small = addFriendInfo.SourceID
				big = addFriendInfo.TargetID
			}
			friendList.IsFriend = true
			friendList.BigID = big
			friendList.SmallID = small
			mysqlQQ.Db.Create(&friendList)
			//判断双方是否在线，若在线，则向对方转发addFriend信息
			if user := core.IOnlineMap.GetUser(addFriend.GetTargetId()); user != nil {
				//当前用户在线
				user.SendMsg(14, &addFriend)
			}
			if user := core.IOnlineMap.GetUser(addFriend.GetSourceId()); user != nil {
				//当前用户在线
				user.SendMsg(14, &addFriend)
			}
		}
		deleteFriendInfo := mysqlQQ.AddFriendList{}
		mysqlQQ.Db.Where("source_id = ?", addFriend.SourceId).Where("target_id = ?", addFriend.TargetId).Delete(&deleteFriendInfo)
		mysqlQQ.Db.Where("source_id = ?", addFriend.TargetId).Where("target_id = ?", addFriend.SourceId).Delete(&deleteFriendInfo)
	}
}
