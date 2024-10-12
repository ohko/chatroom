# Chatroom

chatroom services or embeded chatroom

## screenshot

![screenshot](./screenshot.png)

## api docs

bruno: ./apidoc

## service mode
```shell
$ go run .
- or -
$ go run . -db 'postgres://postgres:postgres@localhost/chatroom?sslmode=disable&TimeZone=Asia/Shanghai'
- or -
$ docker run --rm -it -p 8080:8080 -v /tmp/db/:/db/ ohko/chatroom
```

## embeded mode
```golang
import (
	chatroomCom "github.com/ohko/chatroom/common/com"
	"github.com/ohko/chatroom/srv"
)

if err := chatroomCom.Init("./db/chatroom.db"); err != nil {
	log.Println(err)
}
go srv.PingDeamon()
srv.HandleWS("/ws")
srv.HandleApiFuncs("/api/")
```