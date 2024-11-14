package core

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"sync"
	msg "zinx/mmotest/pb"
	"zinx/ziface"
)

type Player struct {
	Pid  int32              //玩家id
	Conn ziface.IConnection //当前玩家的连接（用于和客户端的连接）
	X    float32            //平面的x坐标
	Y    float32
	Z    float32
	V    float32
}

//PlayerID生成器

var PidGen int32 = 1  //用来生产玩家ID的计数器
var IDLock sync.Mutex //保护PidGen的Mutex

// 创建一个玩家的方法
func NewPlayer(conn ziface.IConnection) *Player {
	//生成一个玩家id
	IDLock.Lock()
	id := PidGen
	PidGen++
	IDLock.Unlock()
	//创建一个玩家对象
	p := Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), //随机在160坐标点，基于x轴若干偏移
		Y:    0,
		Z:    float32(140 + rand.Intn(20)), //随机在140坐标点，基于y轴坐标偏移
		V:    0,
	}
	return &p
}

// 提供一个发送给客户端消息的方法,主要是将pb的protobuf数据序列化之后，再调用zinx的SendMsg方法
func (p *Player) SendMsg(msgId uint32, data proto.Message) {
	//将proto Message结构体序列化 转化成二进制
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("proto message err = ", err)
		return
	}
	//将二进制文件通过zinx的SendMsg将数据发送给客户端
	if p.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}
	if err := p.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Println("Player SendMsg err = ", err)
		return
	}
}

// 告知客户端玩家Pid，同步已经生成的玩家ID给客户端
func (p *Player) SyncPid() {
	//组件MsgID：0的proto数据
	data := &msg.SyncPid{
		Pid: p.Pid,
	}
	//将消息发送给客户端
	p.SendMsg(1, data)
}

// 广播玩家自己的出生地点
func (p *Player) BroadCastStartPosition() {
	//组建MsgID：200的proto数据
	proto_msg := &msg.BroadCast{
		Pid: p.Pid,
		Tp:  2, //Tp2代表广播的位置坐标
		Data: &msg.BroadCast_P{
			P: &msg.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	//将消息发送给客户端
	p.SendMsg(200, proto_msg)
}

// 广播消息的方法
func (p *Player) Talk(content string) {
	//组建一个msgID200proto消息
	proto_msg := &msg.BroadCast{
		Pid: p.Pid,
		Tp:  1, //1代表聊天广播
		Data: &msg.BroadCast_Content{
			Content: content,
		},
	}
	//得到当前世界的所有在线玩家
	players := WorldMgrObj.GetAlPlayers()
	//向所有的玩家发送MsgID200消息
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}
}
