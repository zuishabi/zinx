package goMail

import (
	"gopkg.in/gomail.v2"
)

var Host = "smtp.qq.com"
var Port = 25
var UserName = "861214959@qq.com"
var Password = "iziqiwttyjdfbdef"
var d *gomail.Dialer

func init() {
	d = gomail.NewDialer(Host, Port, UserName, Password)
}

func SendRegisterMail(user string, code string) error {
	message := "您的验证码为："
	m := gomail.NewMessage()
	m.SetHeader("From", UserName)
	m.SetHeader("To", user)
	m.SetHeader("Subject", "GodQQ注册验证码")
	m.SetBody("text/plain", message+code)
	err := d.DialAndSend(m)
	return err
}
