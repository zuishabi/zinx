package utils

import (
	"go.uber.org/zap"
	"io"
	"os"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/utils"
)

//当连接建立时调用的所有函数

func init() {
	core.FunctionLists = append(core.FunctionLists, SendFriendList)
	core.FunctionLists = append(core.FunctionLists, SendAddFriendInfo)
	core.FunctionLists = append(core.FunctionLists, SendChats)
}

// SendAddFriendInfo 向客户端发送全部的好友请求信息
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

// SendFriendList 向客户端发送好友列表
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

// SendChats 向客户端发送所有的历史聊天记录
func SendChats(user *core.User) {
	is_last := false
	chat_list := make([]mysqlQQ.ChatsList, 0)
	for !is_last {
		mysqlQQ.Db.Where("uid = ?", user.Uid).Limit(100).Find(&chat_list)
		if len(chat_list) < 100 {
			is_last = true
		}
		for _, chat := range chat_list {
			if chat.ContentType == 1 {
				msgToClient := msg.MessageToClient{
					Uid:       chat.Friend,
					TargetUid: user.Uid,
					Msg:       &msg.MessageToClient_Text{Text: chat.Content},
					Time:      chat.CreatedAt.Format("2006.01.02.15.04.05"),
					MsgType:   1,
				}
				user.SendMsg(3, &msgToClient)
			} else if chat.ContentType == 2 {
				msgToClient := msg.MessageToClient{
					Uid:       chat.Friend,
					TargetUid: user.Uid,
					Msg:       &msg.MessageToClient_Data{Data: chat.SoundsContent},
					Time:      chat.CreatedAt.Format("2006.01.02.15.04.05"),
					MsgType:   2,
				}
				user.SendMsg(3, &msgToClient)
			} else if chat.ContentType == 3 {
				f, err := os.Open("cache/" + chat.Content + ".png")
				if err != nil {
					utils.L.Error("filed to open image file", zap.Error(err))
					return
				}
				pictureData, err := io.ReadAll(f)
				if err != nil {
					utils.L.Error("failed to read image file", zap.Error(err))
					return
				}
				msgToClient := msg.MessageToClient{
					Uid:       chat.Friend,
					TargetUid: user.Uid,
					Msg:       &msg.MessageToClient_Texture{Texture: &msg.TextureMsg{Data: pictureData}},
					Time:      chat.CreatedAt.Format("2006.01.02.15.04.05"),
					MsgType:   3,
				}
				user.SendMsg(3, &msgToClient)
				if err := f.Close(); err != nil {
					utils.L.Error("close file error", zap.Error(err))
				}
			}
			mysqlQQ.Db.Delete(&chat_list)
		}
	}
}
