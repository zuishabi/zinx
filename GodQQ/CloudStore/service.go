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

// ChunkRequest 用于请求分块
type ChunkRequest struct {
	ChunkID uint64
	FileID  uint64
}

// 存储用户ID以及对应的连接Conn
var connMap sync.Map
var connMapMutex sync.Mutex

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
	err := checkUserConnToCloudStore(fileInfo.UID)
	if err != nil {
		return err
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

// SendUploadFileInfo 向网盘服务器传递交互的信息，例如暂停传输，终止传输，完成传输等
func SendUploadFileInfo(info *gRPCProto.CompleteInfo) error {
	_, err := GRPCClient.UploadFileComplete(context.Background(), info)
	if err != nil {
		return err
	}
	return nil
}

// GetUploadList 获取用户的上传列表
func GetUploadList(uid uint32) []uint64 {
	userInfo := gRPCProto.UserInfo{UID: uid}
	list, err := GRPCClient.GetUploadingFileList(context.Background(), &userInfo)
	if err != nil {
		fmt.Println("获取上传列表失败,err = ", err)
		return nil
	}
	return list.UploadingFilesID
}

// GetUploadedList 获得已经上传的用户的文件列表
func GetUploadedList(uid uint32) *gRPCProto.UploadedFileList {
	userInfo := gRPCProto.UserInfo{UID: uid}
	list, err := GRPCClient.GetUploadedFileList(context.Background(), &userInfo)
	if err != nil {
		return nil
	}
	return list
}

// RequestShareFile 请求新建一个分享
func RequestShareFile(uid uint32, fileID uint64) uint64 {
	shareFile := &gRPCProto.ShareFile{
		UID:    uid,
		FileID: fileID,
	}
	rsp, err := GRPCClient.RequestShareFile(context.Background(), shareFile)
	if err != nil {
		return 0
	}
	return rsp.ShareID
}

// GetShareList 获得已经分享的文件列表
func GetShareList(uid uint32) (*gRPCProto.ShareListRsp, error) {
	r, err := GRPCClient.GetShareList(context.Background(), &gRPCProto.UserInfo{UID: uid})
	return r, err
}

// GetShareFileInfo 通过分享id获得文件的信息
func GetShareFileInfo(shareID uint64) (*gRPCProto.ShareFileInfoRsp, error) {
	r, err := GRPCClient.GetShareFileInfo(context.Background(), &gRPCProto.ShareFileInfo{ShareID: shareID})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// SendFileChunk 向网盘服务器传递数据
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

// DownloadFileChunk 向网盘服务器请求下载一个区块
func DownloadFileChunk(uid uint32, chunkID uint64, fileID uint64) error {
	if err := checkUserConnToCloudStore(uid); err != nil {
		return err
	}
	chunkReq := ChunkRequest{
		ChunkID: chunkID,
		FileID:  fileID,
	}
	value, err := json.Marshal(&chunkReq)
	if err != nil {
		return err
	}
	if err := WriteSendInfo(context.Background(), uid, value); err != nil {
		return err
	}
	return nil
}

// DeleteUploadFile 删除一个已经上传的文件
func DeleteUploadFile(fileID uint64) {
	fileInfo := gRPCProto.FileInfo{FileID: fileID}
	_, _ = GRPCClient.DeleteUploadFile(context.Background(), &fileInfo)
}

func DeleteShareFile(fileID uint64) {
	fileInfo := gRPCProto.FileInfo{FileID: fileID}
	_, _ = GRPCClient.DeleteShareFile(context.Background(), &fileInfo)
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

// ----------------------------------------------------------------------------------------------------------------------
// 检查是否已经创建了与网盘服务的TCP连接，如果没有，则创建
func checkUserConnToCloudStore(uid uint32) error {
	connMapMutex.Lock()
	if _, ok := connMap.Load(uid); !ok {
		connMapMutex.Unlock()
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
		binary.BigEndian.PutUint32(data, uid)
		c.Write(data)
		//等待网盘服务器返回确认
		io.ReadFull(c, data)
		newConn := &Conn{
			TCPConn: c,
			UID:     uid,
			User:    core.IOnlineMap.GetUser(uid),
		}
		connMapMutex.Lock()
		connMap.Store(uid, newConn)
		go startConnReader(newConn)
	}
	connMapMutex.Unlock()
	return nil
}

// 开启tcp通道的读服务，来获得网盘服务器返回的文件数据，同时再返回给用户
func startConnReader(conn *Conn) {
	for {
		idData := make([]byte, 8)
		if _, err := io.ReadFull(conn.TCPConn, idData); err != nil {
			fmt.Println("获取文件下载数据失败,*1 err = ", err)
			break
		}
		lenData := make([]byte, 4)
		if _, err := io.ReadFull(conn.TCPConn, lenData); err != nil {
			fmt.Println("获取文件下载数据失败,*2 err = ", err)
			break
		}
		fileLen := binary.BigEndian.Uint32(lenData)
		data := make([]byte, fileLen)
		if _, err := io.ReadFull(conn.TCPConn, data); err != nil {
			fmt.Println("获取文件下载数据失败,*3 err = ", err)
			break
		}
		// 向用户发送获得的文件数据
		fileID := binary.BigEndian.Uint64(idData)
		downloadData := msg.DownloadChunk{
			FileId: fileID,
			Data:   data,
		}
		conn.User.SendMsg(25, &downloadData)
	}
}
