package main

import (
	"zinx/znet"
)

func main() {
	server := znet.NewServer()
	server.Serve()
}
