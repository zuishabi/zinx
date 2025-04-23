package Routers

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/utils"
	"zinx/ziface"
	"zinx/znet"
)

type SendShareRouter struct {
	znet.BaseRouter
}

func (s *SendShareRouter) Handle(request ziface.IRequest) {
	getShare := &msg.GetShare{}
	proto.Unmarshal(request.GetData(), getShare)
	//传递选择的当前页码的所有信息
	shareList := make([]mysqlQQ.ShareInfo, 5)
	var result *gorm.DB
	var min_id uint
	minShare := mysqlQQ.ShareInfo{}
	sendShare := msg.SendShare{}
	if getShare.Type == 1 {
		//传递全局share
		sendShare.Type = 1
		mysqlQQ.Db.First(&minShare)
		min_id = minShare.ID
		//传递给定页码的全局信息
		result = mysqlQQ.Db.Order("id DESC").Limit(5).Offset((int(getShare.GetPage()) - 1) * 5).Find(&shareList)
	} else if getShare.Type == 2 {
		//传递个人share
		uid, err := request.GetConnection().GetProperty("uid")
		if err != nil {
			utils.L.Error("get user property uid error", zap.Error(err))
			return
		}
		sendShare.Type = 2
		//获得最后一个id
		self := mysqlQQ.Db.Where("uid = ?", uid.(uint32)).Session(&gorm.Session{})
		self.First(&minShare)
		min_id = minShare.ID
		//当是第一页时
		if getShare.Page == 0 {
			result = self.Order("id DESC").Limit(5).Find(&shareList)
		} else {
			result = self.Order("id DESC").Where("id < ?", getShare.GetPage()).Limit(5).Find(&shareList)
		}
	} else {
		return
	}
	if result.Error != nil {
		return
	}
	sendShare.IsTheEnd = false
	for i, share := range shareList {
		//当发送的消息长度大于200时，发送省略信息，并发送索引
		if len(share.Content) >= 200 {
			sendShare.Content = append(sendShare.Content, share.Content[:201])
			sendShare.HideIndex = append(sendShare.HideIndex, uint32(i))
		} else {
			sendShare.Content = append(sendShare.Content, share.Content)
		}
		sendShare.UserId = append(sendShare.UserId, share.Uid)
		sendShare.Time = append(sendShare.Time, share.CreatedAt.Format("2006.01.02.15.04.05"))
		//判断是否是最后一位
		if min_id == share.ID {
			sendShare.IsTheEnd = true
		}
		sendShare.Id = append(sendShare.Id, uint64(share.ID))
	}
	core.IOnlineMap.GetUserByConn(request.GetConnection()).SendMsg(7, &sendShare)
}
