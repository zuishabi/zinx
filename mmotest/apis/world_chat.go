package apis

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/mmotest/core"
	msg "zinx/mmotest/pb"
	"zinx/ziface"
	"zinx/znet"
)

// 世界聊天的路由业务
type WorldChatApi struct {
	znet.BaseRouter
}

func (wc *WorldChatApi) Handle(request ziface.IRequest) {
	//解析客户端传递进来的proto协议
	proto_msg := &msg.Talk{}
	err := proto.Unmarshal(request.GetData(), proto_msg)
	if err != nil {
		fmt.Println("Talk Unmarshal err = ", err)
		return
	}
	//当前的聊天数据是属于哪个玩家发送的
	pid, err := request.GetConnection().GetProperty("pid")
	//根据pid得到对应的player对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//将这个消息广播给其他全部在线的玩家
	player.Talk(proto_msg.Content)
}
