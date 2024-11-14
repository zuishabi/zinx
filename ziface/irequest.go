package ziface

// IRequest接口,实际上是把客户端请求的连接数据和请求的数据包装到了一个request中
type IRequest interface {
	//得到当前连接
	GetConnection() IConnection
	//得到请求的消息数据
	GetData() []byte
	//得到消息的id
	GetMsgID() uint32
}
