package znet

type Message struct {
	ID      uint32 //消息的id
	DataLen uint32 //消息的长度
	Data    []byte //消息的内容
}

// 创建一个message消息包的方法
func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// 获取消息的id
func (m *Message) GetMsgID() uint32 {
	return m.ID
}

// 获取消息的长度
func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

// 获取消息的内容
func (m *Message) GetData() []byte {
	return m.Data
}

// 设置消息的ID
func (m *Message) SetMsgID(id uint32) {
	m.ID = id
}

// 设置消息的内容
func (m *Message) SetData(data []byte) {
	m.Data = data
}

// 设置消息的长度
func (m *Message) SetDataLen(datalen uint32) {
	m.DataLen = datalen
}
