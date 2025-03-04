package CloudStore

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
	gRPCProto "zinx/GodQQ/CloudStore/protocol"
	"zinx/GodQQ/RPC"
)

const (
	GetInfoTopic  = "get_info"  //网关向网盘服务器发送信息
	SendInfoTopic = "send_info" //网盘服务器向网关发送信息
)

var GRPCClient gRPCProto.FilesInfoClient
var GRPCConn *grpc.ClientConn
var TCPAddr string
var GetInfoReader *kafka.Reader
var SendInfoWriter = &kafka.Writer{
	Addr:                   kafka.TCP("127.0.0.1:9092"), //可以传递多个地址来创建多个broker
	Topic:                  GetInfoTopic,
	Balancer:               &kafka.Hash{}, //负载均衡算法，计算哪个partition去哪个broker
	WriteTimeout:           10 * time.Second,
	RequiredAcks:           kafka.RequireOne,
	AllowAutoTopicCreation: true, //是否要自动创建topic
}

// 初始化gRPC客户端
func initCloudStoregRPCClient() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	reply := &RPC.Service{}
	err := RPC.RPCClient.Call(context.Background(), "ServiceManager.GetService", "CloudStoregRPC", &reply)
	if err != nil {
		//未找到合适的服务
		panic(err)
	}
	c, err := grpc.NewClient(fmt.Sprintf("%s:%d", reply.Addr, reply.Port), opts...)
	GRPCConn = c
	GRPCClient = gRPCProto.NewFilesInfoClient(c)
	if err != nil {
		//连接失败
		panic(err)
	}
}

// 初始化kafka读通道以及写通道
func initKafka() {
	go readGetInfo(context.Background())
}

// 从kafka通道中读取数据
func readGetInfo(ctx context.Context) {
	GetInfoReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"127.0.0.1:9092"},
		Topic:          SendInfoTopic,
		CommitInterval: 500 * time.Millisecond,
		StartOffset:    kafka.LastOffset,
		GroupID:        "test",
	})
	for {
		time.Sleep(2 * time.Second)
		fmt.Println("读取数据")
		if message, err := GetInfoReader.ReadMessage(ctx); err != nil {
			fmt.Println("kafka 读取数据失败,error = ", err)
			break
		} else {
			//处理收到的消息，为网盘服务器返回确认收到传递的文件分片
			ProcessGetInfo(&message)
		}
	}
}

func WriteSendInfo(ctx context.Context, key uint32, value []byte) error {
	messageKey := make([]byte, 0)
	binary.BigEndian.PutUint32(messageKey, key)
	if err := SendInfoWriter.WriteMessages(ctx, kafka.Message{Key: messageKey, Value: value}); err != nil {
		fmt.Println("写入kafka失败,error = ", err)
		return err
	}
	return nil
}

// 初始化服务
func InitService() {
	//获得tcp连接的地址
	reply := &RPC.Service{}
	err := RPC.RPCClient.Call(context.Background(), "ServiceManager.GetService", "CloudStoreTCPConn", &reply)
	if err != nil {
		//未找到合适的服务
		panic(err)
	}
	TCPAddr = fmt.Sprintf("%s:%d", reply.Addr, reply.Port)
	initKafka()
	initCloudStoregRPCClient()
}
