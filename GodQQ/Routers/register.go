package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"strconv"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
	"zinx/ziface"
	"zinx/znet"
)

type RegisterRouter struct {
	znet.BaseRouter
}

func (r *RegisterRouter) Handle(request ziface.IRequest) {
	registerMsg := &msg.Register{}
	redis_conn := redisQQ.Pool.Get()
	defer redis_conn.Close()
	err := proto.Unmarshal(request.GetData(), registerMsg)
	if err != nil {
		fmt.Println("[RegisterRouter Handle] : unmarshal register msg err = ", err)
		return
	}
	//检查验证码是否正确
	reply, err := redis_conn.Do("get", "code_"+registerMsg.UserEmail)
	if err != nil {
		fmt.Println("RegisterRouter Handle] : erdis get code err = ", err)
		return
	}
	//验证码不存在
	if reply == nil {
		sendRegisterFailMsg(request.GetConnection(), "验证码不存在")
		return
	}
	if string(reply.([]uint8)) != registerMsg.Code {
		sendRegisterFailMsg(request.GetConnection(), "验证码错误")
		return
	}

	//检查用户名是否重复
	reply, err = redis_conn.Do("sismember", "userName", registerMsg.UserName)
	if err != nil {
		fmt.Println("[RegisterRouter Handle] : redis_conn get userName err = ", err)
		return
	}
	//当redis的set中寻找到此用户名，则用户名重复
	if reply.(int64) == 1 {
		sendRegisterFailMsg(request.GetConnection(), "用户名已存在")
		return
	}

	uid := core.GetUid()
	//在redis中注册邮箱和对应的密码
	redis_conn.Do("set", registerMsg.GetUserEmail(), registerMsg.UserPwd)
	//将用户的uid和用户名存在哈希中
	_, err = redis_conn.Do("hmset", "user_"+registerMsg.GetUserEmail(), "uid", strconv.Itoa(int(uid)), "user_name", registerMsg.GetUserName())
	if err != nil {
		fmt.Println(">>>>>", err)
	}
	//在集合中加入用户名
	redis_conn.Do("sadd", "userName", registerMsg.UserName)
	//删除原来的验证码
	redis_conn.Do("del", "code_"+registerMsg.GetUserEmail())
	sendRegisterSuccessMsg(request.GetConnection())
}

func sendRegisterFailMsg(conn ziface.IConnection, errMsg string) {
	RegisterMsg := &msg.ErrToClient{
		Succ:     false,
		ErrorMsg: errMsg,
	}
	register_msg, err := proto.Marshal(RegisterMsg)
	if err != nil {
		fmt.Println("[RegisterRouter sendRegisterFailMsg] : proto marshal err = ", err)
		return
	}
	err = conn.SendBuffMsg(2, register_msg)
	if err != nil {
		fmt.Println("[RegisterRouter sendRegisterFailMsg] : conn sendMsg err = ", err)
		return
	}
}

func sendRegisterSuccessMsg(conn ziface.IConnection) {
	RegisterMsg := &msg.ErrToClient{
		Succ:     true,
		ErrorMsg: "注册成功",
	}
	register_msg, err := proto.Marshal(RegisterMsg)
	if err != nil {
		fmt.Println("[RegisterRouter sendRegisterSuccessMsg] : proto marshal err = ", err)
		return
	}
	err = conn.SendBuffMsg(2, register_msg)
	if err != nil {
		fmt.Println("[RegisterRouter sendRegisterSuccessMsg] : conn sendMsg err = ", err)
		return
	}
}
