package Routers

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"zinx/GodQQ/mysqlQQ"
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
		fmt.Println("RegisterRouter Handle] : redis get code err = ", err)
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
	user := mysqlQQ.UserInfo{}
	//检查邮箱是否重复
	tx := mysqlQQ.Db.Session(&gorm.Session{SkipDefaultTransaction: true})
	tx.Where("user_email = ?", registerMsg.UserEmail).First(&user)
	if user.UID != 0 {
		//当前邮箱已经被注册
		sendRegisterFailMsg(request.GetConnection(), "邮箱已经被注册")
		return
	}
	//检查用户名是否重复
	tx.Where("user_name = ?", registerMsg.UserName).First(&user)
	if user.UID != 0 {
		sendRegisterFailMsg(request.GetConnection(), "用户名重复")
		return
	}
	//删除原来的验证码
	redis_conn.Do("del", "code_"+registerMsg.GetUserEmail())
	sendRegisterSuccessMsg(request.GetConnection())
	//当前用户名没有被注册，创建新用户
	user.UserName = registerMsg.GetUserName()
	user.Password = registerMsg.GetUserPwd()
	user.UserEmail = registerMsg.GetUserEmail()
	tx.Create(&user)
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
