package ziface

// IDataPack 封包，拆包模块，直接面向TCp连接中的数据流，用于处理tcp粘包问题
type IDataPack interface {
	// GetHeadLen 获取包的长度的方法
	GetHeadLen() uint32
	// Pack 封包方法
	Pack(msg IMessage) ([]byte, error)
	// UnPack 拆包方法
	UnPack([]byte) (IMessage, error)
}
