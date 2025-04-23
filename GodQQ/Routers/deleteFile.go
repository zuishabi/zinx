package Routers

import (
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/CloudStore"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type DeleteUploadFileRouter struct {
	znet.BaseRouter
}

func (d *DeleteUploadFileRouter) Handle(request ziface.IRequest) {
	info := msg.DeleteUploadFile{}
	_ = proto.Unmarshal(request.GetData(), &info)
	CloudStore.DeleteUploadFile(info.FileId)
	//删除完成后进行一次刷新
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	list := CloudStore.GetUploadedList(user.Uid)
	if list == nil {
		return
	}
	rsp := msg.UploadedList{
		FileId:      list.FileID,
		FileLen:     list.FileLen,
		FileName:    list.FileName,
		CreatedTime: list.CreatedTime,
		FileMd5:     list.FileMD5,
	}
	user.SendMsg(21, &rsp)
}

type DeleteShareFileRouter struct {
	znet.BaseRouter
}

func (d *DeleteShareFileRouter) Handle(request ziface.IRequest) {
	info := msg.DeleteShareFile{}
	_ = proto.Unmarshal(request.GetData(), &info)
	CloudStore.DeleteShareFile(info.FileId)
	//删除完成后刷新分享列表
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	rsp, err := CloudStore.GetShareList(user.Uid)
	if err != nil {
		return
	}
	list := &msg.ShareList{
		FileId:      rsp.FileID,
		FileLen:     rsp.FileLen,
		FileName:    rsp.FileName,
		CreatedTime: rsp.CreatedTime,
		ShareId:     rsp.ShareID,
	}
	user.SendMsg(28, list)
}
