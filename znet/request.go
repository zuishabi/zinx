package znet

import "github.com/zuishabi/zinx/ziface"

type Request struct {
	//已经和客户端建立好的连接
	conn ziface.IConnection
	//客户端请求的数据
	msg ziface.IMessage
}

// GetConnection 得到当前连接
func (r *Request) GetConnection() ziface.IConnection {
	return r.conn
}

// GetData 得到请求的消息数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgID()
}
