package Routers

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
	"strconv"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
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
	redis_conn := redisQQ.Pool.Get()
	defer redis_conn.Close()
	reply, err := redis_conn.Do("get", loginMsg.GetUserEmail())
	if err != nil {
		fmt.Println("[LoginRouter Handle] : redis get err = ", err)
		return
	}
	//当返回为空时，表示当前邮箱没有被注册，否则返回用户名对应的密码
	if reply == nil {
		sendLoginFailMsg(request.GetConnection(), "邮件未被注册")
	} else {
		replyString := string(reply.([]uint8))
		if replyString == loginMsg.UserPwd {
			//认证成功，获得用户的id
			reply, err = redis_conn.Do("hget", "user_"+loginMsg.GetUserEmail(), "uid")
			replyString, err = redis.String(reply, err)
			if err != nil {
				fmt.Println("[LoginRouter Handle] : reply to string err = ", err)
				return
			}
			uid, err := strconv.ParseUint(replyString, 10, 32)
			if err != nil {
				fmt.Println("[LoginRouter Handle] : ParseUnit err = ", err)
				return
			}
			_, ok := core.IOnlineMap.UserMap[uint32(uid)]
			if ok {
				//当能在在线列表找到对应的用户，说明已经登录上了
				sendLoginFailMsg(request.GetConnection(), "当前用户已登录")
				return
			}
			reply, err = redis_conn.Do("hget", "user_"+loginMsg.GetUserEmail(), "user_name")
			if err != nil {
				fmt.Println("[LoginRouter Handle] : get user_name err = ", err)
				return
			}
			userName := string(reply.([]uint8))
			user := &core.User{
				Uid:      uint32(uid),
				Conn:     request.GetConnection(),
				UserName: userName,
			}
			request.GetConnection().SetProperty("uid", uint32(uid))
			core.IOnlineMap.AddUser(user)
			sendLoginSuccessMsg(request.GetConnection(), uint32(uid))
		} else {
			sendLoginFailMsg(request.GetConnection(), "密码错误")
		}
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
		UserName: user.UserName,
		Type:     true,
	}
	core.IOnlineMap.BroadCast(5, onOrOffLine)
}
