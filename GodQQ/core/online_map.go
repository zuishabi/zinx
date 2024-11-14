package core

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"sync"
	"zinx/ziface"
)

var IOnlineMap *OnlineMap

type OnlineMap struct {
	UserLock sync.RWMutex
	UserMap  map[uint32]*User
}

func init() {
	IOnlineMap = &OnlineMap{
		UserMap: make(map[uint32]*User),
	}
}

func (m *OnlineMap) AddUser(user *User) {
	m.UserLock.Lock()
	defer m.UserLock.Unlock()
	m.UserMap[user.Uid] = user
	fmt.Println("[OnlineMap AddUser] Add User success")
}

func (m *OnlineMap) RemoveUser(uid uint32) {
	m.UserLock.Lock()
	defer m.UserLock.Unlock()
	delete(m.UserMap, uid)
}

func (m *OnlineMap) GetUser(uid uint32) *User {
	m.UserLock.RLock()
	defer m.UserLock.RUnlock()
	return m.UserMap[uid]
}

func (m *OnlineMap) GetUserByConn(conn ziface.IConnection) *User {
	m.UserLock.RLock()
	defer m.UserLock.RUnlock()
	uid, err := conn.GetProperty("uid")
	if err != nil {
		fmt.Println("[OnlineMap GetUserByConn] get property err = ", err)
		return nil
	}
	return m.UserMap[uid.(uint32)]
}

// 线程安全的获取全部用户
func (m *OnlineMap) GetAllUsers() (users []*User) {
	m.UserLock.RLock()
	defer m.UserLock.RUnlock()
	users = make([]*User, 0)
	for _, user := range m.UserMap {
		users = append(users, user)
	}
	return
}

// 向所有玩家发送消息
func (m *OnlineMap) BroadCast(msgID uint32, message proto.Message) {
	m.UserLock.RLock()
	defer m.UserLock.RUnlock()
	for _, user := range m.UserMap {
		user.SendMsg(msgID, message)
	}
}
