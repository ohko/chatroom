package biz

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ohko/chatroom/config"
)

var (
	clients           sync.Map
	ErrClientNotFound = errors.New("client not found")
)

func WsAddClient(userID int, conn *websocket.Conn) {
	if v, ok := clients.Load(userID); ok {
		v.(*sync.Map).Store(conn, conn)
	} else {
		n := new(sync.Map)
		n.Store(conn, conn)
		clients.Store(userID, n)
	}
}
func WsRemoveClient(userID int) {
	if v, ok := clients.Load(userID); ok {
		v.(*sync.Map).Range(func(key, value any) bool {
			v.(*sync.Map).Delete(key)
			return true
		})
	}
	clients.Delete(userID)
}
func WsRange(f func(key, value any) bool) {
	clients.Range(func(key, value any) bool {
		value.(*sync.Map).Range(f)
		return true
	})
}

func WsSendWelcome(conn *websocket.Conn) error {
	return conn.WriteJSON(config.WSMsg{Type: "connected", CreateTime: time.Now()})
}

func WsSendError(conn *websocket.Conn, err error) error {
	return conn.WriteJSON(config.WSMsg{Type: "error", Data: err.Error()})
}

func WsSendPing(conn *websocket.Conn) error {
	return conn.WriteJSON(config.WSMsg{Type: "ping", CreateTime: time.Now()})
}

func WsSendPong(conn *websocket.Conn) error {
	return conn.WriteJSON(config.WSMsg{Type: "pong", CreateTime: time.Now()})
}

func WsSendBind(conn *websocket.Conn, no int, data any) error {
	return conn.WriteJSON(config.WSMsg{Type: "bind", No: no, Data: data})
}

func WsSendMessageByConn(conn *websocket.Conn, info *config.TableMessage) error {
	return conn.WriteJSON(info)
}

func WsSendMessageByUserID(userID int, info *config.TableMessage) error {
	if v, ok := clients.Load(userID); ok {
		v.(*sync.Map).Range(func(key, value any) bool {
			value.(*websocket.Conn).WriteJSON(info)
			return true
		})
	}
	return ErrClientNotFound
}

func WsNotifyUserGroupJoin(groupID int, joinUserIDs []int) error {
	list, err := UserGroupListByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, l := range list {
		if v, ok := clients.Load(l.UserID); ok {
			v.(*sync.Map).Range(func(key, value any) bool {
				value.(*websocket.Conn).WriteJSON(config.WSMsg{Type: "join", Data: joinUserIDs})
				return true
			})
		}
	}

	return nil
}

func WsNotifyUserGroupRemove(groupID int, removeUserIDs []int) error {
	list, err := UserGroupListByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, l := range list {
		if v, ok := clients.Load(l.UserID); ok {
			v.(*sync.Map).Range(func(key, value any) bool {
				value.(*websocket.Conn).WriteJSON(config.WSMsg{Type: "remove", Data: removeUserIDs})
				return true
			})
		}
	}
	return nil
}
