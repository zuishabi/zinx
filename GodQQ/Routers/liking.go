package Routers

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
	"strconv"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
	"zinx/ziface"
	"zinx/znet"
)

type LikingRouter struct {
	znet.BaseRouter
}

// 点赞或者取消点赞或者查询是否点赞
func (l *LikingRouter) Handle(request ziface.IRequest) {
	likeMsg := msg.Liking{}
	proto.Unmarshal(request.GetData(), &likeMsg)
	if likeMsg.GetContentType() == 0 {
		//先尝试从redis中获取此share的like列表
		getShareLike(&likeMsg)
	}
}

// 查询一个share的like
func getShareLike(likeMsg *msg.Liking) {
	result := msg.Liking{}
	redisConn := redisQQ.Pool.Get()
	defer redisConn.Close()
	key := strconv.Itoa(int(likeMsg.ContentId))
	reply, err := redis.Bool(redisConn.Do("hget", "share_like"+key, likeMsg.GetUserId()))
	if err != nil {
		fmt.Println(err)
		//没有找到此条数据
	} else {
		result.Result = reply
		count, err := redis.Int(redisConn.Do("get", "share_like_counts"+key))
		if err != nil {
			return
		}
		result.Counts = uint32(count)
	}
	fmt.Println(reply)
}
