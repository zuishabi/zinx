package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"strconv"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type SendVideoDataRouter struct {
	znet.BaseRouter
}

func (s *SendVideoDataRouter) Handle(request ziface.IRequest) {
	videoReq := msg.VideoRequest{}
	proto.Unmarshal(request.GetData(), &videoReq)
	//这里获取视频
	startPoint := videoReq.StartPoint
	file, err := os.Open("videos/" + strconv.Itoa(int(videoReq.Id)) + "/" + fmt.Sprintf("%03d", int(startPoint)) + ".mp4")
	if err != nil {
		fmt.Println("file open err = ", err)
		return
	}
	data, err := io.ReadAll(file)
	videoData := msg.VideoData{}
	videoData.Data = data
	videoData.VideoPoint = videoReq.GetStartPoint()
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(17, &videoData)
}
