package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type InquiryRouter struct {
	znet.BaseRouter
}

func (i *InquiryRouter) Handle(request ziface.IRequest) {
	inquiry := &msg.Inquiry{}
	err := proto.Unmarshal(request.GetData(), inquiry)
	if err != nil {
		fmt.Println("[InquiryRouter Handle] unmarshal err = ", err)
		return
	}
	
}
