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
	uid := iUid.(uint32)
	msgToClient := &msg.MessageToClient{
		Uid:  uid,
		Msg:  message.Content,
		Time: time.Now().Format("2006.01.02.15.04.05"),
	}
	core.IOnlineMap.BroadCast(1, msgToClient)
}
