package Routers

import (
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type CreateShareRouter struct {
	znet.BaseRouter
}

func (c *CreateShareRouter) Handle(request ziface.IRequest) {
	createShare := msg.CreateShare{}
	proto.Unmarshal(request.GetData(), &createShare)
	shareInfo := mysqlQQ.ShareInfo{
		Uid:     createShare.UserId,
		Content: createShare.Content,
	}
	tx := mysqlQQ.Db.Session(&gorm.Session{SkipDefaultTransaction: true})
	tx.Create(&shareInfo)
	shareLikeCounts := mysqlQQ.ShareLikeCountsInfo{
		ShareID: shareInfo.ID,
		Counts:  0,
	}
	tx.Create(&shareLikeCounts)
}
