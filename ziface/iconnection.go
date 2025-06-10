package ziface

import "net"

// IConnection 定义连接模块的抽象层
type IConnection interface {
	// Start 启动连接,让当前连接准备开始工作
	Start()
	// Stop 停止连接，结束当前连接的工作
	Stop()
	// GetTCPConnection 获取当前连接的绑定socket conn
	GetTCPConnection() *net.TCPConn
	// GetConnID 获取当前链接模块的链接id
	GetConnID() uint32
	// RemoteAddr 获取远程客户端的TCP状态，IP port
	RemoteAddr() net.Addr
	// SendMsg 发送数据，将数据发送给远程的客户端
	SendMsg(uint32, []byte) error
	// SendBuffMsg 直接将Message数据发送给远程的TCP客户端(有缓冲)
	SendBuffMsg(msgId uint32, data []byte) error
	// SetProperty 设置连接属性
	SetProperty(key string, value interface{})
	// GetProperty 获取连接属性
	GetProperty(key string) (interface{}, error)
	// RemoveProperty 移除连接属性
	RemoveProperty(ley string)
}
