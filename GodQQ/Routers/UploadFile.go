package Routers

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"zinx/GodQQ/CloudStore"
	gRPCProto "zinx/GodQQ/CloudStore/protocol"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type UploadFileListReqRouter struct {
	znet.BaseRouter
}

// Handle 请求获得上传列表
func (u *UploadFileListReqRouter) Handle(request ziface.IRequest) {
	fmt.Println("获得上传列表")
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	fileList := CloudStore.GetUploadList(user.Uid)
	fmt.Println(fileList)
	rep := msg.UploadList{FileId: fileList}
	user.SendMsg(20, &rep)
}

type UploadFileReqRouter struct {
	znet.BaseRouter
}

func (u *UploadFileReqRouter) Handle(request ziface.IRequest) {
	uploadFileReq := msg.UploadReq{}
	_ = proto.Unmarshal(request.GetData(), &uploadFileReq)
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	uploadFileInfo := gRPCProto.UploadFileInfo{
		UID:      user.Uid,
		FileName: uploadFileReq.FileName,
		MD5:      uploadFileReq.Md5,
		FileID:   uploadFileReq.FileId,
		FileLen:  uploadFileReq.FileLen,
	}
	//发送一个上传文件的请求，会创建对应的服务
	err := CloudStore.UploadFileReq(&uploadFileInfo, uploadFileReq.ClientId)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("传输成功")
}

type UploadFileChunkRouter struct {
	znet.BaseRouter
}

func (u *UploadFileChunkRouter) Handle(request ziface.IRequest) {
	uploadFileChunk := msg.UploadChunk{}
	_ = proto.Unmarshal(request.GetData(), &uploadFileChunk)
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	CloudStore.SendFileChunk(user.Uid, uploadFileChunk.FileId, uploadFileChunk.Data)
}

type UploadFileInfoRouter struct {
	znet.BaseRouter
}

func (u *UploadFileInfoRouter) Handle(request ziface.IRequest) {
	uploadFileInfo := msg.UploadInfo{}
	_ = proto.Unmarshal(request.GetData(), &uploadFileInfo)
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	info := gRPCProto.CompleteInfo{
		UID:    user.Uid,
		FileID: uploadFileInfo.FileId,
	}
	if uploadFileInfo.Type == 1 {
		info.Complete = 1
		//为暂停传输
	} else if uploadFileInfo.Type == 2 {
		//为终止传输
		info.Complete = 2
	} else if uploadFileInfo.Type == 3 {
		//为完成传输
		info.Complete = 3
	}
	err := CloudStore.SendUploadFileInfo(&info)
	if err != nil {
		fmt.Println("传递文件的完成，暂停，终止信息失败")
		return
	}
}

type UploadedFileListRouter struct {
	znet.BaseRouter
}

func (u *UploadedFileListRouter) Handle(request ziface.IRequest) {
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
	}
	user.SendMsg(21, &rsp)
}
