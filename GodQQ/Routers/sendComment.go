package Routers

import (
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/ziface"
	"zinx/znet"
)

type SendCommentRouter struct {
	znet.BaseRouter
}

func (s *SendCommentRouter) Handle(request ziface.IRequest) {
	req := msg.GetComment{}
	proto.Unmarshal(request.GetData(), &req)
	shareComments := make([]mysqlQQ.ShareComment, 5)
	var minShare mysqlQQ.ShareComment
	currentShare := mysqlQQ.Db.Where("share_id = ?", req.GetId()).Session(&gorm.Session{})
	currentShare.First(&minShare)
	if req.GetPage() == 0 {
		//申请开始的五条数据
		currentShare.Order("id DESC").Limit(5).Find(&shareComments)
	} else {
		currentShare.Order("id DESC").Where("id < ?", req.Page).Limit(5).Find(&shareComments)
	}
	userIDList := make([]uint32, 0)
	comments := make([]string, 0)
	createTimes := make([]string, 0)
	commentIDs := make([]uint64, 0)
	flag := false //判断是否是最后一个
	for _, v := range shareComments {
		userIDList = append(userIDList, v.Uid)
		comments = append(comments, v.Content)
		createTimes = append(createTimes, v.CreatedAt.Format("2006.01.02.15.04.05"))
		commentIDs = append(commentIDs, uint64(v.ID))
		if v.ID == minShare.ID {
			flag = true
		}
	}
	tmsg := msg.SendComment{
		Id:          req.GetId(),
		UserId:      userIDList,
		Comment:     comments,
		CommentTime: createTimes,
		CommentId:   commentIDs,
		IsTheEnd:    flag,
	}
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(11, &tmsg)
}
