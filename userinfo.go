package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// UserInfo type
type UserInfo struct {
	uid       int
	bnickname []byte
	nickname  string
	conn      net.Conn
}

// UserSyncMap type
type UserSyncMap struct {
	sync.Map
}

// ToString override for logging
func (user UserInfo) ToString() string {
	return fmt.Sprintf("[%d,%s]", user.uid, user.nickname)
}

func (user UserInfo) writeln(b []byte) (bool, error) {
	_, err := user.conn.Write(b)
	if err != nil {
		return false, err
	}

	_, err = user.conn.Write([]byte(outputLineTerm))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (user UserInfo) swriteln(s string) {
	user.writeln([]byte(s))
}

func (user *UserInfo) updateNickname(n string) {
	user.nickname = n
	user.bnickname = []byte(n + nickMsgDelimiter)
	user.swriteln("Your nickname has been changed to " + user.nickname)
}

func (m *UserSyncMap) createNewUser(uid int, c net.Conn) *UserInfo {
	user := &UserInfo{
		uid:  uid,
		conn: c,
	}
	user.updateNickname(fmt.Sprintf("GUEST%03d", uid))

	m.Store(uid, user)

	return user
}

func (m *UserSyncMap) onlineUserList() string {

	nicks := []string{}

	m.Range(func(key interface{}, user interface{}) bool {
		nicks = append(nicks, user.(*UserInfo).nickname)
		return true
	})

	return "ONLINE NOW: " + strings.Join(nicks, ",")
}

func (m *UserSyncMap) disconnectUser(user *UserInfo, reason string) {
	m.Delete(user.uid)

	go func() {
		defer user.conn.Close()
		user.swriteln(fmt.Sprintf("The server has disconnected you. %s", reason))
		time.Sleep(2 * time.Second)
	}()
}
func (m *UserSyncMap) disconnectUsersWithNickname(nickname string) {

	m.Range(func(key interface{}, user interface{}) bool {
		if user.(*UserInfo).nickname == nickname {
			m.disconnectUser(user.(*UserInfo), "Another user has registered this nickname.")
		}
		return true
	})
}

func (m *UserSyncMap) dispatchUserMessage(mi MsgInfo) {

	fromUserI, ok := m.Load(mi.uid)
	if !ok {
		//user not found online
		return
	}

	m.Range(func(key interface{}, user interface{}) bool {

		if mi.uid != user.(*UserInfo).uid {
			user.(*UserInfo).writeln(append(fromUserI.(*UserInfo).bnickname, mi.msg...))
		}
		return true
	})
}

func isValidNickname(n string) bool {
	return len(n) >= minNickLen
}
