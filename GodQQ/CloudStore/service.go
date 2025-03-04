package CloudStore

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io"
	"net"
	"sync"
	gRPCProto "zinx/GodQQ/CloudStore/protocol"
	"zinx/GodQQ/core"
	msg "zinx/GodQQ/protocol"
)

type Conn struct {
	TCPConn *net.TCPConn
	UID     uint32     //对应用户的uid
	User    *core.User //对应的用户
}

type ChunkAck struct {
	ChunkID uint64
	UserID  uint32
	FileID  uint64
	Error   error
}

// 存储用户ID以及对应的连接Conn
var connMap sync.Map

// ProcessGetInfo 处理由网盘服务器发送的确认消息
func ProcessGetInfo(message *kafka.Message) {
	//解析出用户ID，同时获得对应用户
	uid := binary.BigEndian.Uint32(message.Key)
	user := core.IOnlineMap.GetUser(uid)
	//向用户发出继续传输的信息
	chunkAck := ChunkAck{}
	json.Unmarshal(message.Value, &chunkAck)
	uploadFileChunk := msg.UploadChunk{
		Chunk:  chunkAck.ChunkID + 1,
		FileId: chunkAck.FileID,
	}
	user.SendMsg(24, &uploadFileChunk)
	fmt.Println("传输中...")
}

func UploadFileReq(fileInfo *gRPCProto.UploadFileInfo, clientID uint32) error {
	// 检查是否已经创建了与网盘服务的TCP连接，如果没有，则创建
	if _, ok := connMap.Load(fileInfo.UID); !ok {
		addr, err := net.ResolveTCPAddr("tcp", TCPAddr)
		if err != nil {
			return err
		}
		c, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			return err
		}
		//向tcp通道写入自己的ID
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, fileInfo.UID)
		c.Write(data)
		//等待网盘服务器返回确认
		io.ReadFull(c, data)
		newConn := &Conn{
			TCPConn: c,
			UID:     fileInfo.UID,
			User:    core.IOnlineMap.GetUser(fileInfo.UID),
		}
		connMap.Store(fileInfo.UID, newConn)
	}
	rsp, err := GRPCClient.RequestUploadFile(context.Background(), fileInfo)
	if err != nil {
		return err
	}
	user := core.IOnlineMap.GetUser(fileInfo.UID)
	//向客户端写回文件信息
	uploadInfo := msg.UploadInfo{
		Type:     0,
		FileId:   rsp.FileID,
		ClientId: clientID,
	}
	user.SendMsg(26, &uploadInfo)
	uploadFileChunk := msg.UploadChunk{
		Chunk:  rsp.ChunkID,
		FileId: rsp.FileID,
	}
	user.SendMsg(24, &uploadFileChunk)
	return nil
}

func SendFileChunk(Uid uint32, FileID uint64, Data []byte) {
	c, ok := connMap.Load(Uid)
	if !ok {
		fmt.Println("未找到用户")
	}
	conn := c.(*Conn).TCPConn
	idData := make([]byte, 8)
	binary.BigEndian.PutUint64(idData, FileID)
	conn.Write(idData)
	//写入数据的长度
	lenData := make([]byte, 4)
	binary.BigEndian.PutUint32(lenData, uint32(len(Data)))
	conn.Write(lenData)
	//写入数据
	conn.Write(Data)
	fmt.Println("写入成功")
}

// SendUploadFileInfo 向网盘服务器传递交互的信息，例如暂停传输，终止传输，完成传输等
func SendUploadFileInfo(info *gRPCProto.CompleteInfo) error {
	_, err := GRPCClient.UploadFileComplete(context.Background(), info)
	if err != nil {
		return err
	}
	return nil
}

func CloseConn(uid uint32) {
	c, ok := connMap.Load(uid)
	if !ok {
		return
	}
	conn := c.(*Conn)
	conn.TCPConn.Close()
	connMap.Delete(uid)
}
