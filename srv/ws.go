package srv

import (
	"encoding/json"
	"log"
	"net/http"
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
	hookAfterRecvMessage HookAfterRecvMessage
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	if err := biz.WsSendWelcome(conn); err != nil {
		log.Println(err)
		return
	}

	fromUserID := 0

	for {
		_, bs, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg config.WSMsg
		if err = json.Unmarshal(bs, &msg); err != nil {
			if err := biz.WsSendError(conn, err); err != nil {
				return
			}
			continue
		}

		switch msg.Type {
		case "ping":
			biz.WsSendPong(conn)
		case "bind":
			if token, err := deToken(msg.Token); err == nil {
				fromUserID = token.UserID
				biz.WsAddClient(token.UserID, conn)
				biz.WsSendBind(conn, 0, "success")
			} else {
				biz.WsSendBind(conn, 1, "failed")
			}
		case "text", "image":
			msg.FromUserID = fromUserID
			info := config.TableMessage{
				Type:          msg.Type,
				FromUserID:    msg.FromUserID,
				ToUserID:      msg.ToUserID,
				GroupID:       msg.GroupID,
				Content:       msg.Content,
				CreateTime:    time.Now(),
				SenderExtData: msg.SenderExtData,
				ExtData:       msg.ExtData,
			}
			selfInfo := info
			if err := biz.MessageSend(&info); err != nil {
				biz.WsSendError(conn, err)
			}
			// notify self
			selfInfo.FromUser = info.FromUser
			biz.WsSendMessageByConn(conn, &selfInfo)
			msg.MessageID = info.MessageID
		}

		msg.CreateTime = time.Now()
		if hookAfterRecvMessage != nil {
			hookAfterRecvMessage(msg)
		}
	}
}

func PingDeamon() {
	for {
		time.Sleep(time.Second * 30)
		biz.WsRange(func(key, value any) bool {
			if err := biz.WsSendPing(value.(*websocket.Conn)); err != nil {
				biz.WsRemoveClient(key.(int))
				value.(*websocket.Conn).Close()
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

type HookAfterRecvMessage func(msg config.WSMsg)
