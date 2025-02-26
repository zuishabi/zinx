package RPC

import (
	"context"
	"github.com/zuishabi/ZRPC/Client"
	"log"
	"time"
)

type Service struct {
	Addr string
	Port uint32
	Name string
}

var RPCClient *Client.Client

func InitClient() {
	var err error
	RPCClient, err = Client.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		panic(err)
		return
	}
	service := Service{
		Addr: "127.0.0.1",
		Port: 8999,
		Name: "MainServer",
	}
	var reply int
	err = RPCClient.Call(context.Background(), "ServiceManager.RegisterService", service, &reply)
	if err != nil {
		panic(err)
		return
	}
	log.Printf("Service registered with reply: %d", reply)
	go heatBeat()
}

func heatBeat() {
	var reply int
	service := Service{
		Addr: "127.0.0.1",
		Port: 8999,
		Name: "MainServer",
	}
	for {
		select {
		case <-time.After(time.Second * 2):
			err := RPCClient.Call(context.Background(), "ServiceManager.CheckHeartBeat", service, &reply)
			if err != nil {
				log.Println("heartbeat error = ", err)
				return
			}
		}
	}
}
