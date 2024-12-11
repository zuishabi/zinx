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

type BroadCastRouter struct {
	znet.BaseRouter
}

func (b *BroadCastRouter) Handle(request ziface.IRequest) {
	message := &msg.MessageFromClient{}
	err := proto.Unmarshal(request.GetData(), message)
	if err != nil {
		fmt.Println("[BroadCastRouter Handle] : unmarshal msg err = ", err)
		return
	}
	iUid, err := request.GetConnection().GetProperty("uid")
	if err != nil {
		fmt.Println("[BroadCastRouter Handle] : get property uid err = ", err)
		return
	}
	var msgToClient *msg.MessageToClient
	uid := iUid.(uint32)
	if message.MsgType == 1 {
		msgToClient = &msg.MessageToClient{
			Uid:     uid,
			Msg:     &msg.MessageToClient_Text{Text: message.GetText()},
			MsgType: message.MsgType,
			Time:    time.Now().Format("2006.01.02.15.04.05"),
		}
	} else if message.MsgType == 2 {
		msgToClient = &msg.MessageToClient{
			Uid:     uid,
			Msg:     &msg.MessageToClient_Data{Data: message.GetData()},
			MsgType: message.MsgType,
			Time:    time.Now().Format("2006.01.02.15.04.05"),
		}
	} else if message.MsgType == 3 {
		//传递图片，记录了图片的长宽以及数据
		msgToClient = &msg.MessageToClient{
			Uid: uid,
			Msg: &msg.MessageToClient_Texture{Texture: &msg.TextureMsg{
				Width:  message.GetTexture().Width,
				Height: message.GetTexture().Height,
				Data:   message.GetTexture().GetData(),
				Format: message.GetTexture().GetFormat(),
			}},
			MsgType: message.MsgType,
			Time:    time.Now().Format("2006.01.02.15.04.05"),
		}
	}
	core.IOnlineMap.BroadCast(1, msgToClient)
}
