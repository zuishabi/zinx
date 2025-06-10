package ziface

// IMessage 将请求的消息封装到一个Message中，定义抽象的接口
type IMessage interface {
	// GetMsgID 获取消息的id
	GetMsgID() uint32
	// GetDataLen 获取消息的长度
	GetDataLen() uint32
	// GetData 获取消息的内容
	GetData() []byte
	// SetMsgID 设置消息的ID
	SetMsgID(uint32)
	// SetData 设置消息的内容
	SetData([]byte)
	// SetDataLen 设置消息的长度
	SetDataLen(uint32)
}
