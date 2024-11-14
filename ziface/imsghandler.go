package ziface

//消息管理抽象层

type IMsgHandler interface {
	// DoMsgHandler 调度对应的Router消息处理方法
	DoMsgHandler(IRequest)
	// AddRouter 为消息添加具体的处理路由
	AddRouter(msgID uint32, router IRouter)
	// StartWorkerPool 启动Worker工作池
	StartWorkerPool()
	// SendMsgToTaskQueue 将消息发送给消息队列，从而传递给worker进行业务处理
	SendMsgToTaskQueue(IRequest)
}
