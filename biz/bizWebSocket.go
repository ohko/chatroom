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
	clients.Store(userID, conn)
}
func WsRemoveClient(userID int) {
	clients.Delete(userID)
}
func WsRange(f func(key, value any) bool) {
	clients.Range(f)
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
	if conn, ok := clients.Load(userID); ok {
		return conn.(*websocket.Conn).WriteJSON(info)
	}
	return ErrClientNotFound
}

func WsNotifyUserGroupJoin(groupID int, joinUserIDs []int) error {
	list, err := UserGroupListByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, l := range list {
		if conn, ok := clients.Load(l.UserID); ok {
			conn.(*websocket.Conn).WriteJSON(config.WSMsg{Type: "join", Data: joinUserIDs})
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
		if conn, ok := clients.Load(l.UserID); ok {
			conn.(*websocket.Conn).WriteJSON(config.WSMsg{Type: "remove", Data: removeUserIDs})
		}
	}
	return nil
}
