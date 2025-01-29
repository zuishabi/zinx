package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"time"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type PrivateChatRouter struct {
	znet.BaseRouter
}

func (p *PrivateChatRouter) Handle(request ziface.IRequest) {
	msgFromClient := &msg.MessageFromClient{}
	err := proto.Unmarshal(request.GetData(), msgFromClient)
	if err != nil {
		fmt.Println("[PrivateChatRouter Handle] proto unmarshal err = ", err)
		return
	}
	uid, err := request.GetConnection().GetProperty("uid")
	if err != nil {
		fmt.Println("[PrivateChatRouter Handle] get property uid err = ", err)
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
	}
	core.IOnlineMap.GetUser(uid.(uint32)).SendMsg(3, msgToClient)
}
