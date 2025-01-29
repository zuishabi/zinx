package routers

import (
	"google.golang.org/protobuf/proto"
	video_msg "zinx/goVideo/proto"
	"zinx/goVideo/utils"
	"zinx/ziface"
	"zinx/znet"
)

type SendVideoInfoRouter struct {
	znet.BaseRouter
}

// 获取对应视频的信息
func (s *SendVideoInfoRouter) Handle(request ziface.IRequest) {
	videoInfoReq := video_msg.VideoInfo{}
	proto.Unmarshal(request.GetData(), &videoInfoReq)
	url := "video_cache/1/1.mp4"
	len := utils.GetLength(url)
	videoInfoReq.VideoLen = len
	data, _ := proto.Marshal(&videoInfoReq)
	request.GetConnection().SendBuffMsg(2, data)
}
