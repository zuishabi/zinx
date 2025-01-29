package routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"strconv"
	video_msg "zinx/goVideo/proto"
	"zinx/ziface"
	"zinx/znet"
)

type SendVideoRouter struct {
	znet.BaseRouter
}

func (s *SendVideoRouter) Handle(request ziface.IRequest) {
	videoReq := video_msg.VideoRequest{}
	proto.Unmarshal(request.GetData(), &videoReq)
	//这里获取视频
	if videoReq.VideoId == 1 {
		startPoint := videoReq.StartPoint
		file, err := os.Open("video_cache/1/1.mp4" + "_" + strconv.Itoa(int(startPoint)) + ".mp4")
		if err != nil {
			fmt.Println("file open err = ", err)
			return
		}
		data, err := io.ReadAll(file)
		videoData := video_msg.VideoData{}
		videoData.Data = data
		videoData.VideoPoint = videoReq.GetStartPoint()
		sendData, _ := proto.Marshal(&videoData)
		request.GetConnection().SendBuffMsg(3, sendData)
	}
}
