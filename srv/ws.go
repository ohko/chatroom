package srv

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ohko/chatroom/biz"
	"github.com/ohko/chatroom/config"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var clients sync.Map

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	fromUserID := 0

	for {
		_, bs, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg WSMsg
		if err = json.Unmarshal(bs, &msg); err != nil {
			log.Println(err)
			continue
		}

		switch msg.Type {
		case "ping":
			conn.WriteJSON(WSMsg{Type: "pong"})
		case "bind":
			if token, err := checkToken(msg.Token); err == nil {
				fromUserID = token.UserID
				clients.Store(token.UserID, conn)
			}
			conn.WriteJSON(WSMsg{Type: "bind", No: 0, Data: "success"})
		case "text":
			info := config.TableMessage{
				Type:       msg.Type,
				FromUserID: fromUserID,
				ToUserID:   msg.ToUserID,
				GroupID:    msg.GroupID,
				Content:    msg.Content,
				CreateTime: time.Now(),
			}
			if err := biz.MessageSend(&info); err != nil {
				conn.WriteJSON(WSMsg{Type: "text", No: 1, Data: err.Error()})
				break
			}
			msg.FromUserID = fromUserID
			msg.MessageID = info.MessageID
			if msg.GroupID != 0 {
				userGroups, err := biz.UserGroupListByGroupID(msg.GroupID)
				if err != nil {
					conn.WriteJSON(WSMsg{Type: "text", No: 1, Data: err.Error()})
					break
				}
				for _, ug := range userGroups {
					if toConn, ok := clients.Load(ug.UserID); ok {
						toConn.(*websocket.Conn).WriteJSON(msg)
					}
				}
			} else {
				if toConn, ok := clients.Load(msg.ToUserID); ok {
					toConn.(*websocket.Conn).WriteJSON(msg)
				}
			}
		}
	}
}

func PingDeamon() {
	for {
		time.Sleep(time.Second * 15)
		clients.Range(func(key, value any) bool {
			if err := value.(*websocket.Conn).WriteJSON(WSMsg{Type: "ping"}); err != nil {
				clients.Delete(key)
			}
			return true
		})
	}
}

func HandleWS() {
	http.HandleFunc("/ws", wsHandler)
}

type WSMsg struct {
	Type       string // ping/pong/text/image
	Token      string `json:",omitempty"` // type=bind
	MessageID  int    `json:",omitempty"`
	FromUserID int    `json:",omitempty"`
	ToUserID   int    `json:",omitempty"`
	GroupID    int    `json:",omitempty"`
	Content    string `json:",omitempty"`
	No         int    `json:",omitempty"`
	Data       string `json:",omitempty"`
}
