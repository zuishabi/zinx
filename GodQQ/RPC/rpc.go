package RPC

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

type Service struct {
	Addr string
	Port uint32
	Name string
}

var RPCClient *rpc.Client

func InitClient() {
	var err error
	RPCClient, err = rpc.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		panic(err)
		return
	}
	service := Service{
		Addr: "122.228.237.118",
		Port: 17868,
		Name: "MainServer",
	}
	var reply int
	err = RPCClient.Call("ServiceManager.RegisterService", service, &reply)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Service registered with reply: %d", reply)
	go heatBeat()
}

func heatBeat() {
	var reply int
	service := Service{
		Addr: "122.228.237.118",
		Port: 17868,
		Name: "MainServer",
	}
	for {
		select {
		case <-time.After(time.Second * 2):
			err := RPCClient.Call("ServiceManager.CheckHeartBeat", service, &reply)
			if err != nil {
				log.Println("heartbeat error = ", err)
				return
			}
		}
	}
}
