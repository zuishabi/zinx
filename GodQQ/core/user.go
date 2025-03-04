package core

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/ziface"
)

var FunctionLists []func(user *User)

// 用户的结构体
type User struct {
	UserName string
	Uid      uint32
	Conn     ziface.IConnection
}

func (u *User) SendMsg(msgId uint32, data proto.Message) {
	//将proto Message结构体序列化 转化成二进制
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("proto message err = ", err)
		return
	}
	//当用户退出后
	if u == nil {
		fmt.Println("用户已退出", msgId)
		return
	}
	//将二进制文件通过zinx的SendMsg将数据发送给客户端
	if u.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}
	if err := u.Conn.SendBuffMsg(msgId, msg); err != nil {
		fmt.Println("Player SendMsg err = ", err)
		return
	}
}

func (u *User) OnUserReady() {
	for _, i := range FunctionLists {
		i(u)
	}
}
