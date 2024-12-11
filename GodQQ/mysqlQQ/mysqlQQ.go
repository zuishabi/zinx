package mysqlQQ

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"zinx/GodQQ/core"
)

var dsn string = "root:861214959@tcp(127.0.0.1:3306)/game?charset=utf8mb4&parseTime=True&loc=Local"
var Db *gorm.DB

func Start() error {
	var err error
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db, err := Db.DB()
	fmt.Println(db.Stats().MaxOpenConnections)
	fmt.Println(db.Stats().MaxIdleTimeClosed)
	fmt.Println(db.Stats().MaxIdleClosed)
	fmt.Println(db.Stats().MaxLifetimeClosed)
	if err != nil {
		return err
	}
	//当IsUpdated时，更新所有的数据库
	if core.ConfigObj.IsUpdate {
		MigrateDatabase()
	}
	return nil
}

// 将对象自动迁移到数据库中
func MigrateDatabase() {
	Db.AutoMigrate(&ShareInfo{})
	Db.AutoMigrate(&ShareComment{})
	Db.AutoMigrate(&UserInfo{})
	Db.AutoMigrate(&ShareLikeInfo{})
	Db.AutoMigrate(&ShareCommentsLikeInfo{})
	Db.AutoMigrate(&ShareCommentsLikeCountsInfo{})
	Db.AutoMigrate(&ShareLikeCountsInfo{})
}
