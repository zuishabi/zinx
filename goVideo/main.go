package main

import (
	"zinx/goVideo/routers"
	"zinx/goVideo/utils"
	"zinx/znet"
)

func main() {
	//os.Mkdir("video_cache", os.ModePerm)
	//os.MkdirAll("video_cache/1", os.ModePerm)
	utils.InitVideo("video_cache/1/1.mp4")
	server := znet.NewServer()
	server.AddRouter(1, &routers.SendVideoRouter{})
	server.AddRouter(2, &routers.SendVideoInfoRouter{})
	server.Serve()
}
