package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/CloudStore"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type DownloadFileRouter struct {
	znet.BaseRouter
}

func (d *DownloadFileRouter) Handle(request ziface.IRequest) {
	req := msg.DownloadChunk{}
	_ = proto.Unmarshal(request.GetData(), &req)
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	if err := CloudStore.DownloadFileChunk(user.Uid, req.ChunkId, req.FileId); err != nil {
		fmt.Println("-!-", err)
	}
}
