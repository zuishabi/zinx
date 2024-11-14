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
	msgToClient := &msg.MessageToClient{
		Uid:       uid.(uint32),
		TargetUid: msgFromClient.Uid,
		Msg:       msgFromClient.Content,
		Time:      time.Now().Format("2006.01.02.15.04.05"),
	}
	core.IOnlineMap.GetUser(msgFromClient.Uid).SendMsg(3, msgToClient)
	core.IOnlineMap.GetUser(uid.(uint32)).SendMsg(3, msgToClient)
}
