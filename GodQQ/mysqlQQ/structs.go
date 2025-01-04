package mysqlQQ

import (
	"time"
)

type ShareInfo struct {
	ID        uint `gorm:"primarykey;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Uid       uint32
	Content   string
}

type ShareComment struct {
	ID        uint `gorm:"primarykey;uniqueIndex;autoIncrement"`
	ShareID   uint //所处share的id
	CreatedAt time.Time
	Uid       uint32 //创建者的uid
	TargetUid uint32 //评论的目标
	Content   string
}

type UserInfo struct {
	UID       uint32 `gorm:"primarykey;uniqueIndex;autoIncrement"`
	CreatedAt time.Time
	Password  string
	UserName  string `gorm:"uniqueIndex;size:10"`
	UserEmail string `gorm:"Index"`
}

// share的喜欢表
type ShareLikeInfo struct {
	ShareID uint   `gorm:"Index"`
	UserID  uint32 `gorm:"Index"`
	IsLike  bool
}

// ShareComment的喜欢表
type ShareCommentsLikeInfo struct {
	CommentID uint   `gorm:"Index"`
	UserID    uint32 `gorm:"Index"`
	IsLike    bool
}

// sharelike的数量表
type ShareLikeCountsInfo struct {
	ShareID uint `gorm:"primarykey;uniqueIndex"`
	Counts  uint32
}

// share comment的数量表
type ShareCommentsLikeCountsInfo struct {
	ShareCommentID uint `gorm:"primarykey;uniqueIndex"`
	Counts         uint32
}
