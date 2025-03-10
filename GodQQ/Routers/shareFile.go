package Routers

import (
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/CloudStore"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type RequestShareFileRouter struct {
	znet.BaseRouter
}

func (r *RequestShareFileRouter) Handle(request ziface.IRequest) {
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	shareReq := msg.RequestShareFile{}
	_ = proto.Unmarshal(request.GetData(), &shareReq)
	shareRsp := CloudStore.RequestShareFile(user.Uid, shareReq.FileId)
	rsp := msg.RequestShareFileRsp{ShareId: shareRsp}
	user.SendMsg(27, &rsp)
}

type GetShareListRouter struct {
	znet.BaseRouter
}

func (g *GetShareListRouter) Handle(request ziface.IRequest) {
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

type GetShareFileInfo struct {
	znet.BaseRouter
}

func (g *GetShareFileInfo) Handle(request ziface.IRequest) {
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	req := msg.GetShareFileInfo{}
	_ = proto.Unmarshal(request.GetData(), &req)
	shareFileInfoRsp := msg.GetShareFileInfoRsp{}
	rsp, err := CloudStore.GetShareFileInfo(req.ShareId)
	if err != nil {
		shareFileInfoRsp.ShareId = 0
		user.SendMsg(29, &shareFileInfoRsp)
		return
	}
	shareFileInfoRsp.ShareId = req.ShareId
	shareFileInfoRsp.FileName = rsp.FileName
	shareFileInfoRsp.FileLen = rsp.FileLen
	shareFileInfoRsp.FileId = rsp.FileID
	shareFileInfoRsp.FileMd5 = rsp.FileMD5
	user.SendMsg(29, &shareFileInfoRsp)
}
