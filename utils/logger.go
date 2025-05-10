package utils

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

const Topic = "Log"

var L *zap.Logger
var CTX context.Context
var SendInfoWriter = &kafka.Writer{
	Addr:                   kafka.TCP("127.0.0.1:9094"), //可以传递多个地址来创建多个broker
	Topic:                  Topic,
	Balancer:               &kafka.Hash{}, //负载均衡算法，计算哪个partition去哪个broker
	WriteTimeout:           10 * time.Second,
	RequiredAcks:           kafka.RequireOne,
	AllowAutoTopicCreation: true, //是否要自动创建topic
}

type logEncoder struct {
	zapcore.Encoder
}

type kafkaWriter struct{}

func (k *kafkaWriter) Write(p []byte) (n int, err error) {
	err = SendInfoWriter.WriteMessages(context.Background(), kafka.Message{Key: []byte("MainServer"), Value: p})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// InitLogger 初始化日志系统，如果mode为0则为development，否则会传输到kafka中
func InitLogger(mode int) {
	if mode == 0 {
		L, _ = zap.NewDevelopment()
		return
	} else {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := logEncoder{Encoder: zapcore.NewJSONEncoder(cfg.EncoderConfig)}
		core := zapcore.NewCore(encoder, zapcore.AddSync(&kafkaWriter{}), zap.InfoLevel)
		L = zap.New(core, zap.AddCaller())
	}
}
