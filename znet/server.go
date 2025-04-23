package znet

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

// IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器的名称
	Name string
	//服务器的ip版本
	IPVersion string
	//服务器监听的ip
	IP string
	//服务器监听的端口
	Port int
	//当前的Server的消息管理模块，用来绑定MsgID和对应的处理业务
	MsgHandler ziface.IMsgHandler
	//该server的连接管理器
	ConnManager ziface.IConnManager
	//该Server创建连接之后自动调用Hook函数
	OnConnStart func(conn ziface.IConnection)
	//该Server销毁连接之后自动调用Hook函数
	OnConnStop func(conn ziface.IConnection)
}

func (s *Server) Start() {
	fmt.Printf("[Zinx] Server Name: %s, listener at IP: %s,Port: %d is starting\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] Version: %s,MaxConn: %d,MaxPackageSize: %d\n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPackageSize)
	go func() {
		//开启消息队列及工作池
		s.MsgHandler.StartWorkerPool()
		//获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			utils.L.Error("resolve tcp addr error", zap.Error(err))
			return
		}
		//监听服务器的地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			utils.L.Error("listen tcp error", zap.Error(err))
			return
		}
		var cid uint32 = 0
		//阻塞等待客户端连接，处理客户端连接业务(读写)
		for {
			//如果有客户端连接过来，阻塞会返回
			conn, err := listener.AcceptTCP()
			if err != nil {
				utils.L.Error("accept tcp error", zap.Error(err))
				continue
			}
			//设置最大连接个数的判断，如果超过最大连接，那么则关闭此新的连接
			if s.ConnManager.Len() >= utils.GlobalObject.MaxConn {
				utils.L.Error("too many conn", zap.Error(errors.New("用户过多")))
				conn.Close()
				continue
			}
			//将该处理新链接的业务方法和conn进行绑定，得到我们的连接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++
			//启动当前的连接业务
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	//将一些服务器的资源，状态或者一些已经开辟的连接信息进行停止或者回收
	fmt.Println("[STOP]Zinx server name ", s.Name)
	s.ConnManager.ClearConn()

}

func (s *Server) Serve() {
	//启动server的服务业务
	s.Start()

	//设置阻塞
	select {}
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnManager
}

// 注册OnConnStart钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		s.OnConnStop(conn)
	}
}

func NewServer() ziface.IServer {
	s := &Server{
		Name:        utils.GlobalObject.Name,
		IPVersion:   "tcp",
		IP:          utils.GlobalObject.Host,
		Port:        utils.GlobalObject.TcpPort,
		MsgHandler:  NewHandler(),
		ConnManager: NewConnManager(),
	}
	return s
}
