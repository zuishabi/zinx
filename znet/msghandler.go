package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

// MsgHandler 消息处理模块的实现
type MsgHandler struct {
	//存放每个msgID所对应的处理方法
	Apis map[uint32]ziface.IRouter
	//负责worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//业务工作Worker池的worker数量
	WorkerPoolSize uint32
}

// NewHandler 初始化、创建MsgHandler的方法
func NewHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, //从全局配置中获取
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

// DoMsgHandler 调度对应的Router消息处理方法
func (mh *MsgHandler) DoMsgHandler(request ziface.IRequest) {
	//从Request中找到msgID
	router, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgID(), "is NOT FOUND!NEED REGISTER")
		return
	}
	//根据msgid调度对应router业务即可
	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)
}

// AddRouter 为消息添加具体的处理路由
func (mh *MsgHandler) AddRouter(msgID uint32, router ziface.IRouter) {
	//判断当前绑定的api处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//id已经注册了
		panic("repeat api ,msgID = " + strconv.Itoa(int(msgID)))
	} else {
		//添加msgID和router的绑定关系
		mh.Apis[msgID] = router
		fmt.Println("add api success")
	}

}

// StartWorkerPool 启动一个worker工作池(开启工作池的动作只能发生一次，一个zinx框架只能有一个worker工作池)
func (mh *MsgHandler) StartWorkerPool() {
	//根据workerPoolSize分别开启Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//当前的worker对应的channel消息队列 开辟空间 第0个worker就用第0个channel
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前的worker,阻塞等待消息从channel中传递来
		go mh.startOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个worker工作流程
func (mh *MsgHandler) startOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	//不断阻塞等待对应消息队列的消息
	for {
		select {
		//如果有消息过来，出列的就是一个客户端的request，执行当前request所绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// SendMsgToTaskQueue 将消息交给TaskQueue，由Worker进行处理
func (mh *MsgHandler) SendMsgToTaskQueue(request ziface.IRequest) {
	//将消息平均分配给不同的worker
	//根据客户端建立的connid来进行分配
	//基本的平均分配的轮询
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize

	//将消息发送给对应的worker的taskqueue即可
	mh.TaskQueue[workerID] <- request
}
