package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type SendVideoRouter struct {
	znet.BaseRouter
}

func (s *SendVideoRouter) Handle(request ziface.IRequest) {
	getRequest := msg.GetVideoList{}
	err := proto.Unmarshal(request.GetData(), &getRequest)
	if err != nil {
		fmt.Println("proto unmarshal error = ", err)
		return
	}
	videoInfoList := make([]mysqlQQ.VideoList, 0)
	if getRequest.Page == 0 {
		//查找的是第一页
		mysqlQQ.Db.Order("id DESC").Limit(6).Find(&videoInfoList)
	} else {
		mysqlQQ.Db.Order("id DESC").Where("id < ?", getRequest.Page).Limit(6).Find(&videoInfoList)
	}
	lastVideoInfo := mysqlQQ.VideoList{}
	mysqlQQ.Db.Last(&lastVideoInfo)
	sendVideoInfoList := msg.SendVideoList{}
	sendVideoInfoList.IsLast = false
	for _, v := range videoInfoList {
		sendVideoInfoList.VideoId = append(sendVideoInfoList.VideoId, v.ID)
		sendVideoInfoList.VideoName = append(sendVideoInfoList.VideoName, v.VideoName)
		sendVideoInfoList.VideoLen = append(sendVideoInfoList.VideoLen, v.VideoLen)
		sendVideoInfoList.VideoDescription = append(sendVideoInfoList.VideoDescription, v.VideoDescription)
		sendVideoInfoList.VideoCreateTime = append(sendVideoInfoList.VideoCreateTime, v.CreatedAt.Format("2006.01.02.15.04.05"))
		sendVideoInfoList.VideoPlayTime = append(sendVideoInfoList.VideoPlayTime, v.PlayTime)
		if v.ID == lastVideoInfo.ID {
			sendVideoInfoList.IsLast = true
		}
	}
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(16, &sendVideoInfoList)
}
