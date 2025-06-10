package znet

import (
	"errors"
	"github.com/zuishabi/zinx/utils"
	"github.com/zuishabi/zinx/ziface"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//当前连接的关闭状态
	isClosed bool
	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandler
	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	//有关冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte
	//当前心跳更新的时间
	heartbeatTime time.Time
	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

// NewConnection 创建连接的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandler) *Connection {
	//初始化Conn属性
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}
	//将新创建的Conn添加到链接管理中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				utils.L.Warn("Send Data error,Conn Writer exit", zap.Error(err))
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					utils.L.Warn("Send Buff Data error,Conn Writer exit", zap.Error(err))
					return
				}
			} else {
				break
			}
		case <-c.ExitBuffChan:
			return
		}
	}
}

// StartReader 启动读携程
func (c *Connection) StartReader() {
	utils.L.Info("Conn connect", zap.Uint32("id", 1))
	defer c.Stop()
	for {
		// 创建拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			utils.L.Warn("read msg_head error", zap.Error(err))
			break
		}

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			utils.L.Warn("unpack error ", zap.Error(err))
			break
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				utils.L.Warn("read msg data error ", zap.Error(err))
				break
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		// 更新当前连接的心跳状态
		c.UpdateHeartbeat()

		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// StartCheckHeartbeat 开始检查心跳，超过5分钟没有进行心跳更新的关闭连接
func (c *Connection) StartCheckHeartbeat() {
	for {
		select {
		case <-time.After(time.Second * 10):
			c.propertyLock.RLock()
			if time.Since(c.heartbeatTime) > time.Minute*5 {
				// 长时间没有发送心跳包，断开连接
				c.propertyLock.RUnlock()
				c.Stop()
				return
			}
			c.propertyLock.RUnlock()
		}
	}
}

// Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	// 检查是否需要开启心跳检测
	if utils.GlobalObject.UseHeartbeat {
		c.heartbeatTime = time.Now()
		go c.StartCheckHeartbeat()
	}
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)
}

// Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	utils.L.Info("Conn disconnect", zap.Uint32("id", 2))
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	_ = c.Conn.Close()
	//关闭Writer
	c.ExitBuffChan <- true

	//将链接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c)

	//关闭该链接全部管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

// GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// SendMsg 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgChan <- msg

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		return errors.New("Pack error msg ")
	}
	select {
	case <-idleTimeout.C:
		return errors.New("send buff msg timeout")
	case c.msgBuffChan <- msg:
		return nil
	}
}

// SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

// UpdateHeartbeat 更新当前心跳状态
func (c *Connection) UpdateHeartbeat() {
	c.propertyLock.Lock()
	c.heartbeatTime = time.Now()
	c.propertyLock.Unlock()
}
