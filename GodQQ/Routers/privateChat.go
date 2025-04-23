package Routers

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"os"
	"time"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/utils"
	"zinx/ziface"
	"zinx/znet"
)

type PrivateChatRouter struct {
	znet.BaseRouter
}

func (p *PrivateChatRouter) Handle(request ziface.IRequest) {
	msgFromClient := &msg.MessageFromClient{}
	_ = proto.Unmarshal(request.GetData(), msgFromClient)
	uid, err := request.GetConnection().GetProperty("uid")
	if err != nil {
		utils.L.Error("get property uid error", zap.Error(err))
		return
	}
	//TODO先判断是否是好友

	//发送消息
	var msgToClient *msg.MessageToClient
	if msgFromClient.MsgType == 1 {
		msgToClient = &msg.MessageToClient{
			Uid:       uid.(uint32),
			Msg:       &msg.MessageToClient_Text{Text: msgFromClient.GetText()},
			MsgType:   msgFromClient.MsgType,
			Time:      time.Now().Format("2006.01.02.15.04.05"),
			TargetUid: msgFromClient.Uid,
		}
	} else if msgFromClient.MsgType == 2 {
		msgToClient = &msg.MessageToClient{
			Uid:       uid.(uint32),
			Msg:       &msg.MessageToClient_Data{Data: msgFromClient.GetData()},
			MsgType:   msgFromClient.MsgType,
			Time:      time.Now().Format("2006.01.02.15.04.05"),
			TargetUid: msgFromClient.Uid,
		}
	} else if msgFromClient.MsgType == 3 {
		msgToClient = &msg.MessageToClient{
			Uid: uid.(uint32),
			Msg: &msg.MessageToClient_Texture{Texture: &msg.TextureMsg{
				Width:  msgFromClient.GetTexture().Width,
				Height: msgFromClient.GetTexture().Height,
				Data:   msgFromClient.GetTexture().Data,
				Format: msgFromClient.GetTexture().Format,
			}},
			MsgType:   msgFromClient.MsgType,
			Time:      time.Now().Format("2006.01.02.15.04.05"),
			TargetUid: msgFromClient.Uid,
		}
	}
	//判断对方是否在线
	if user := core.IOnlineMap.GetUser(msgFromClient.Uid); user != nil {
		//当前用户在线
		user.SendMsg(3, msgToClient)
	} else {
		//TODO当前用户不在线
		fmt.Println("对方不在线")
		chat := mysqlQQ.ChatsList{}
		chat.UID = msgToClient.TargetUid
		chat.Friend = msgToClient.Uid
		chat.ContentType = msgToClient.MsgType
		if chat.ContentType == 1 {
			//如果是文字消息
			chat.Content = msgToClient.GetText()
		} else if chat.ContentType == 2 {
			//如果是语音信息
			chat.SoundsContent = msgToClient.GetData()
		} else if chat.ContentType == 3 {
			//如果是图片信息
			//将图片保存到文件中
			chat.Content = utils.BytesMD5(msgToClient.GetTexture().GetData())
			err := os.WriteFile("cache/"+chat.Content+".png", msgToClient.GetTexture().GetData(), 0644)
			if err != nil {
				utils.L.Error("save picture to the cache error", zap.Error(err))
				return
			}
		}
		mysqlQQ.Db.Create(&chat)
	}
	core.IOnlineMap.GetUser(uid.(uint32)).SendMsg(3, msgToClient)
}
