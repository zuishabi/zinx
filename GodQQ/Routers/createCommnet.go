package Routers

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/utils"
	"zinx/ziface"
	"zinx/znet"
)

type CreateCommentRouter struct {
	znet.BaseRouter
}

func (c *CreateCommentRouter) Handle(request ziface.IRequest) {
	createComment := msg.CreateComment{}
	proto.Unmarshal(request.GetData(), &createComment)
	uid, err := request.GetConnection().GetProperty("uid")
	if err != nil {
		utils.L.Error("get property uid error", zap.Error(err))
		return
	}
	shareComment := mysqlQQ.ShareComment{
		ShareID:   uint(createComment.ShareId),
		TargetUid: createComment.GetTargetUid(),
		Uid:       uid.(uint32),
		Content:   createComment.Content,
	}
	tx := mysqlQQ.Db.Session(&gorm.Session{SkipDefaultTransaction: true})
	tx.Create(&shareComment)
	shareCommentsCounts := mysqlQQ.ShareCommentsLikeCountsInfo{
		ShareCommentID: shareComment.ID,
		Counts:         0,
	}
	tx.Create(&shareCommentsCounts)
}
