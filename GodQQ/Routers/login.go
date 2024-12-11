package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type LoginRouter struct {
	znet.BaseRouter
}

func (l *LoginRouter) Handle(request ziface.IRequest) {
	loginMsg := &msg.LoginFromClient{}
	err := proto.Unmarshal(request.GetData(), loginMsg)
	if err != nil {
		fmt.Println("[LoginRouter Handle] : unmarshal request err = ", err)
		return
	}
	//先检查邮箱是否正确
	user := mysqlQQ.UserInfo{}
	mysqlQQ.Db.Where("user_email = ?", loginMsg.GetUserEmail()).First(&user)
	if user.UID == 0 {
		sendLoginFailMsg(request.GetConnection(), "用户未注册")
		return
	}
	if user.Password == loginMsg.UserPwd {
		//密码正确
		//如果当前用户已经登录
		core.IOnlineMap.UserLock.RLock()
		_, ok := core.IOnlineMap.UserMap[user.UID]
		if ok {
			//当能在在线列表找到对应的用户，说明已经登录上了
			sendLoginFailMsg(request.GetConnection(), "当前用户已登录")
			core.IOnlineMap.UserLock.RUnlock()
			return
		}
		core.IOnlineMap.UserLock.RUnlock()
		iuser := core.User{
			Uid:      user.UID,
			Conn:     request.GetConnection(),
			UserName: user.UserName,
		}
		request.GetConnection().SetProperty("uid", user.UID)
		core.IOnlineMap.AddUser(&iuser)
		sendLoginSuccessMsg(request.GetConnection(), user.UID)
	} else {
		//密码错误
		sendLoginFailMsg(request.GetConnection(), "密码错误")
	}
}

// 向客户端发送登录失败的消息
func sendLoginFailMsg(conn ziface.IConnection, err_msg string) {
	loginMsg := &msg.ErrToClient{
		Succ:     false,
		ErrorMsg: err_msg,
	}
	login_msg, err := proto.Marshal(loginMsg)
	if err != nil {
		fmt.Println("[LoginRouter sendLoginFailMsg] : proto marshal err = ", err)
		return
	}
	err = conn.SendBuffMsg(0, login_msg)
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
