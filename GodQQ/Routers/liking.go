package Routers

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"strconv"
	"time"
	"zinx/GodQQ/core"
	"zinx/GodQQ/mysqlQQ"
	msg "zinx/GodQQ/protocol"
	"zinx/GodQQ/redisQQ"
	"zinx/ziface"
	"zinx/znet"
)

/*
	2025/1/2
*/

type LikingRouter struct {
	znet.BaseRouter
}

func init() {
	go aggregateLikes()
}

// 聚合写入点赞数量
func aggregateLikes() {
	redisConn := redisQQ.Pool.Get()
	defer redisConn.Close()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 使用 SCAN 命令查找所有以 aggregate_share_counts 为前缀的键
			var cursor int64
			var keys []string
			for {
				reply, err := redis.Values(redisConn.Do("SCAN", cursor, "MATCH", "aggregate_share_counts*"))
				if err != nil {
					fmt.Println("SCAN error:", err)
					break
				}
				reply, _ = redis.Scan(reply, &cursor, &keys)
				if cursor == 0 {
					break
				}
			}
			// 聚合所有键的值并写入数据库
			for _, key := range keys {
				count, err := redis.Int(redisConn.Do("GET", key))
				if err != nil {
					fmt.Println("GET error:", err)
					continue
				}
				// 从键中提取 share_id
				shareIDStr := key[len("aggregate_share_counts"):]
				shareID, err := strconv.Atoi(shareIDStr)
				if err != nil {
					fmt.Println("strconv.Atoi error:", err)
					continue
				}
				// 更新数据库中的点赞数量
				shareLikeCountsInfo := mysqlQQ.ShareLikeCountsInfo{}
				mysqlQQ.Db.Where("share_id = ?", shareID).First(&shareLikeCountsInfo)
				if count > 0 {
					shareLikeCountsInfo.Counts += uint32(count)
				} else {
					shareLikeCountsInfo.Counts -= uint32(-count)
				}
				mysqlQQ.Db.Save(&shareLikeCountsInfo)
				// 删除 Redis 中的键
				redisConn.Do("DEL", key)
			}
		}
	}
}

// 点赞或者取消点赞或者查询是否点赞
func (l *LikingRouter) Handle(request ziface.IRequest) {
	likeMsg := msg.Liking{}
	user := core.IOnlineMap.GetUserByConn(request.GetConnection())
	proto.Unmarshal(request.GetData(), &likeMsg)
	if likeMsg.GetContentType() == 1 {
		//先尝试从redis中获取此share的like列表
		getShareLike(&likeMsg, user)
	} else if likeMsg.GetContentType() == 2 {
		//修改like
		setShareLike(&likeMsg, user)
	}
}

// 查询一个share的like
func getShareLike(likeMsg *msg.Liking, user *core.User) {
	result := msg.Liking{}
	result.ContentId = likeMsg.ContentId
	result.ContentType = likeMsg.ContentType
	result.UserId = likeMsg.UserId
	redisConn := redisQQ.Pool.Get()
	defer redisConn.Close()
	//将shareid作为关键字存储起来，通过不同的前缀加上关键字在redis中查找数据
	key := strconv.Itoa(int(likeMsg.ContentId))
	reply, err := redis.Int(redisConn.Do("get", "share_like_counts"+key))
	if err != nil {
		fmt.Println("getShareLike err=", err)
		fmt.Println(redisConn.Do("get", "share_like_counts"+key))
		//没有找到此条数据，从数据库中寻找这条数据
		fmt.Println("getShareLike,没有找到此条数据，从数据库中寻找这条数据")
		inquiryCounts := mysqlQQ.ShareLikeCountsInfo{}
		tx := mysqlQQ.Db.Session(&gorm.Session{SkipDefaultTransaction: true})
		inquiryUser := mysqlQQ.ShareLikeInfo{}
		tx.Where("share_id = ?", likeMsg.GetContentId()).First(&inquiryCounts)
		tx.Where("share_id = ?", likeMsg.GetContentId()).Where("user_id = ?", likeMsg.GetUserId()).First(&inquiryUser)
		result.Counts = inquiryCounts.Counts
		result.Result = inquiryUser.IsLike
		//将数据库中的数据放到redis中
		_, err := redisConn.Do("set", "share_like_counts"+key, inquiryCounts.Counts)
		if err != nil {
			fmt.Println("set_share_like_counts err = ", err)
			return
		}
		//将所有的点赞者都放到redis中
		//TODO 限制放入redis中数据的长度
		userList := make([]mysqlQQ.ShareLikeInfo, 0)
		tx.Where("share_id = ?", likeMsg.GetContentId()).Where("is_like = ?", true).Find(&userList)
		//将从数据库中获取的所有的点赞用户数据和点赞量存入redis中
		redisConn.Send("multi")
		for _, v := range userList {
			redisConn.Send("sadd", "share_like"+key, v.UserID)
		}
		redisConn.Send("expire", "share_like_counts"+key, 600)
		redisConn.Send("expire", "share_like"+key, 600)
		_, err = redisConn.Do("exec")
		if err != nil {
			fmt.Println("multi error = ", err)
			return
		}
	} else {
		//当在redis缓存中查到了数据，查询对应的是否点赞以及点赞的数量,同时重新开始计时
		result.Counts = uint32(reply)
		reply, _ := redis.Bool(redisConn.Do("sismember", "share_like"+key, user.Uid))
		result.Result = reply
		redisConn.Do("expire", "share_like_counts"+key, 600)
		redisConn.Do("expire", "share_like"+key, 600)
		//TODO 当在redis中查到了需要查找的内容id但是没有当前用户的信息时，再去数据库查找
	}
	user.SendMsg(12, &result)
}

func setShareLike(likeMsg *msg.Liking, user *core.User) {
	redisConn := redisQQ.Pool.Get()
	defer redisConn.Close()

	// 使用事务处理数据库和 Redis 更新
	tx := mysqlQQ.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Println("Transaction rollback due to panic:", r)
		}
	}()

	// 将用户点赞信息写入 MySQL 中
	info := mysqlQQ.ShareLikeInfo{}
	result := tx.Where("share_id = ?", likeMsg.ContentId).Where("user_id = ?", likeMsg.UserId).First(&info)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 记录不存在，创建新记录
			info = mysqlQQ.ShareLikeInfo{
				ShareID: uint(likeMsg.ContentId),
				UserID:  likeMsg.UserId,
				IsLike:  likeMsg.Result,
			}
			if err := tx.Create(&info).Error; err != nil {
				tx.Rollback()
				fmt.Println("Create error:", err)
				return
			}
		} else {
			// 处理其他错误
			tx.Rollback()
			fmt.Println("Error:", result.Error)
			return
		}
	} else {
		// 当数据库中存储的喜欢和发送过来的喜欢数据一样，则直接返回
		if likeMsg.Result == info.IsLike {
			tx.Rollback()
			return
		}
		info.IsLike = likeMsg.Result
		if err := tx.Model(&info).Where("share_id = ?", info.ShareID).Where("user_id = ?", info.UserID).Update("IsLike", likeMsg.Result).Error; err != nil {
			tx.Rollback()
			fmt.Println("Update error:", err)
			return
		}
	}

	// 检查 Redis 中是否已经有数据
	key := strconv.Itoa(int(likeMsg.ContentId))
	currentCount, err := redis.Int(redisConn.Do("get", "share_like_counts"+key))
	if err != nil {
		fmt.Println("setShareLike err=", err)
		// 在 Redis 中没有数据, 将从数据库中获取的所有的点赞用户数据和点赞量存入 Redis 中, 并设置过期时间
		shareLikeCountsInfo := mysqlQQ.ShareLikeCountsInfo{}
		if err := tx.Where("share_id = ?", likeMsg.ContentId).First(&shareLikeCountsInfo).Error; err != nil {
			tx.Rollback()
			fmt.Println("Query error:", err)
			return
		}
		fmt.Println("set,没有在redis中找到对应数据，获取mysql中数据shareLikeCountsInfo:", shareLikeCountsInfo)
		// 直接将查询到的数据在数据库和缓存中进行更新
		if likeMsg.Result == true {
			shareLikeCountsInfo.Counts += 1
			fmt.Println("set,点赞，数量：", shareLikeCountsInfo.Counts)
		} else {
			shareLikeCountsInfo.Counts -= 1
			fmt.Println("set,取消点赞，数量：", shareLikeCountsInfo.Counts)
		}
		if err := tx.Save(&shareLikeCountsInfo).Error; err != nil {
			tx.Rollback()
			fmt.Println("Save error:", err)
			return
		}
		if _, err := redisConn.Do("set", "share_like_counts"+key, shareLikeCountsInfo.Counts); err != nil {
			tx.Rollback()
			fmt.Println("Redis set error:", err)
			return
		}
		userList := make([]mysqlQQ.ShareLikeInfo, 0)
		if err := tx.Where("share_id = ?", likeMsg.GetContentId()).Where("is_like = ?", true).Find(&userList).Error; err != nil {
			tx.Rollback()
			fmt.Println("Query user list error:", err)
			return
		}
		fmt.Println("set,寻找到所有喜欢列表:", userList)
		redisConn.Send("multi")
		for _, v := range userList {
			redisConn.Send("sadd", "share_like"+key, v.UserID)
		}
		redisConn.Send("expire", "share_like_counts"+key, 600)
		redisConn.Send("expire", "share_like"+key, 600)
		if _, err := redisConn.Do("exec"); err != nil {
			tx.Rollback()
			fmt.Println("Redis exec error:", err)
			return
		}
	} else {
		if likeMsg.Result == true {
			currentCount += 1
			redisConn.Do("sadd", "share_like"+key, likeMsg.UserId)
		} else {
			currentCount -= 1
			_, err2 := redisConn.Do("srem", "share_like"+key, likeMsg.UserId)
			if err2 != nil {
				fmt.Println("设置错误！！！！！！！！！！！！！！！！！！", err2)
				return
			}
		}
		if _, err := redisConn.Do("set", "share_like_counts"+key, currentCount); err != nil {
			tx.Rollback()
			fmt.Println("Redis set error:", err)
			return
		}
		redisConn.Do("expire", "share_like_counts"+key, 600)
		redisConn.Do("expire", "share_like"+key, 600)
		// 将修改的点赞数量再放入Redis中，来进行聚合写入
		aggregateCount, err := redis.Int(redisConn.Do("get", "aggregate_share_counts"+key))
		if err != nil {
			aggregateCount = 0
		}
		if likeMsg.Result == true {
			aggregateCount += 1
		} else {
			aggregateCount -= 1
		}
		if _, err := redisConn.Do("set", "aggregate_share_counts"+key, aggregateCount); err != nil {
			tx.Rollback()
			fmt.Println("Redis set aggregate error:", err)
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		fmt.Println("Transaction commit error:", err)
	}
}
