package Routers

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
	"zinx/utils"
	"zinx/ziface"
	"zinx/znet"
)

type LoginRouter struct {
	znet.BaseRouter
}

func (l *LoginRouter) Handle(request ziface.IRequest) {
	redisConn := redisQQ.Pool.Get()
	defer redisConn.Close()
	loginMsg := &msg.LoginFromClient{}
	_ = proto.Unmarshal(request.GetData(), loginMsg)
	//从客户端获得的key中在redis中查找是否存在
	uid, err := redis.Int(redisConn.Do("get", loginMsg.Key))
	if err != nil {
		utils.L.Error("user login error,get login key error", zap.Error(err))
		sendLoginFailMsg(request.GetConnection(), "登录失败")
		return
	}
	redisConn.Do("del", loginMsg.Key)
	//查找当前用户是否在线
	core.IOnlineMap.UserLock.RLock()
	_, ok := core.IOnlineMap.UserMap[uint32(uid)]
	if ok {
		//当能在在线列表找到对应的用户，说明已经登录上了
		sendLoginFailMsg(request.GetConnection(), "当前用户已登录")
		core.IOnlineMap.UserLock.RUnlock()
		return
	}
	core.IOnlineMap.UserLock.RUnlock()
	userInfo := mysqlQQ.UserInfo{}
	mysqlQQ.Db.Where("uid = ?", uid).First(&userInfo)
	//创建用户到在线列表
	iuser := core.User{
		Uid:      uint32(uid),
		Conn:     request.GetConnection(),
		UserName: userInfo.UserName,
	}
	request.GetConnection().SetProperty("uid", uint32(uid))
	core.IOnlineMap.AddUser(&iuser)
	sendLoginSuccessMsg(request.GetConnection(), uint32(uid))
}

// 向客户端发送登录失败的消息
func sendLoginFailMsg(conn ziface.IConnection, err_msg string) {
	loginMsg := &msg.ErrToClient{
		Succ:     false,
		ErrorMsg: err_msg,
	}
	m, _ := proto.Marshal(loginMsg)
	err := conn.SendBuffMsg(0, m)
	if err != nil {
		fmt.Println("[LoginRouter sendLoginFailMsg] : SendMsg err = ", err)
	}
}

// 向客户端发送登录成功的消息
func sendLoginSuccessMsg(conn ziface.IConnection, uid uint32) {
	loginMsg := &msg.ErrToClient{
		Succ:     true,
		ErrorMsg: "登录成功",
		Info:     &msg.ErrToClient_Uid{Uid: uid},
	}
	login_msg, err := proto.Marshal(loginMsg)
	if err != nil {
		fmt.Println("[LoginRouter sendLoginSuccessMsg] : proto marshal err = ", err)
		return
	}
	err = conn.SendBuffMsg(0, login_msg)
	if err != nil {
		fmt.Println("[LoginRouter sendLoginSuccessMsg] : SendMsg err = ", err)
	}
	//向玩家广播登录消息
	user := core.IOnlineMap.GetUserByConn(conn)
	onOrOffLine := &msg.OnOrOffLineMsg{
		Uid:      user.Uid,
		Type:     true,
		UserName: user.UserName,
	}
	core.IOnlineMap.BroadCast(5, onOrOffLine)
}
