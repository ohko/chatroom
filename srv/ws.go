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

var (
	clients              sync.Map
	hookAfterRecvMessage HookAfterRecvMessage
)

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
			if token, err := deToken(msg.Token); err == nil {
				fromUserID = token.UserID
				clients.Store(token.UserID, conn)
			}
			conn.WriteJSON(WSMsg{Type: "bind", No: 0, Data: "success"})
		case "text":
			msg.FromUserID = fromUserID
			info := config.TableMessage{
				Type:       msg.Type,
				FromUserID: msg.FromUserID,
				ToUserID:   msg.ToUserID,
				GroupID:    msg.GroupID,
				Content:    msg.Content,
			}
			if err := SendMessage(&info); err != nil {
				conn.WriteJSON(WSMsg{Type: "text", No: 1, Data: err.Error()})
			}
			msg.MessageID = info.MessageID
		case "addGroup": // TODO
		case "online": // TODO
		}

		msg.CreateTime = time.Now()
		if hookAfterRecvMessage != nil {
			hookAfterRecvMessage(msg)
		}
	}
}

func SendMessage(info *config.TableMessage) error {
	return biz.MessageSend(info, func(userID int, info *config.TableMessage) {
		if toConn, ok := clients.Load(userID); ok {
			toConn.(*websocket.Conn).WriteJSON(info)
		}
	})
}

func PingDeamon() {
	for {
		time.Sleep(time.Second * 30)
		clients.Range(func(key, value any) bool {
			if err := value.(*websocket.Conn).WriteJSON(WSMsg{Type: "ping"}); err != nil {
				clients.Delete(key)
			}
			return true
		})
	}
}

func HandleWS(path string) {
	http.HandleFunc(path, wsHandler)
}

func SetHookAfterRecvMessage(fun HookAfterRecvMessage) {
	hookAfterRecvMessage = fun
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
	CreateTime time.Time
}

type HookAfterRecvMessage func(msg WSMsg)
