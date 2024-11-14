package Routers

import (
	"fmt"
	"github.com/badoux/checkmail"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"strings"
	"zinx/GodQQ/goMail"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
	"zinx/ziface"
	"zinx/znet"
)

var codeLen = 6
var codeList = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

type GenerateCaptchaRouter struct {
	znet.BaseRouter
}

func (g *GenerateCaptchaRouter) Handle(request ziface.IRequest) {
	redisConn := redisQQ.Pool.Get()
	defer redisConn.Close()
	emailAddress := string(request.GetData())
	err := checkmail.ValidateFormat(emailAddress)
	if err != nil {
		sendErrMsg("邮箱格式错误", false, request.GetConnection())
		return
	}
	err = checkmail.ValidateHost(emailAddress)
	if err != nil {
		sendErrMsg("无法解析的邮箱", false, request.GetConnection())
		return
	}
	//err = checkmail.ValidateHostAndUser(goMail.Host, goMail.UserName, emailAddress)
	//if err != nil {
	//	sendErrMsg("访问的邮箱不存在", false, request.GetConnection())
	//	return
	//}
	//检查邮件地址是否已经被注册
	reply, err := redisConn.Do("get", emailAddress)
	if err != nil {
		fmt.Println("redis获取邮箱失败", err)
		return
	}
	if reply != nil {
		sendErrMsg("邮箱已被注册", false, request.GetConnection())
		return
	}
	sendErrMsg("验证码发送成功", true, request.GetConnection())
	code := getCode()
	err = goMail.SendRegisterMail(emailAddress, code)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = redisConn.Do("setex", "code_"+emailAddress, 300, code)
	if err != nil {
		fmt.Println("redis设置验证码错误,err = ", err)
		return
	}
}

func sendErrMsg(errMsg string, flag bool, conn ziface.IConnection) {
	RegisterMsg := &msg.ErrToClient{}
	RegisterMsg.ErrorMsg = errMsg
	RegisterMsg.Succ = flag
	msgData, err := proto.Marshal(RegisterMsg)
	if err != nil {
		fmt.Println("解析验证邮箱信息失败")
		return
	}
	conn.SendBuffMsg(6, msgData)
}

// 获得验证码
func getCode() string {
	r := len(codeList)
	var sb strings.Builder
	for i := 0; i < codeLen; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", codeList[rand.Intn(r)])
	}
	return sb.String()
}
