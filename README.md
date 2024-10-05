# Chatroom

chatroom services or embeded chatroom

## screenshot

![screenshot](./screenshot.png)

## api docs

https://doc.apipost.net/docs/34ab26f038e5000

## service mode
```shell
$ go run .
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