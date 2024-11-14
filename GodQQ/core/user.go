package core

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/ziface"
)

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
