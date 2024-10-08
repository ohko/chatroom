package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/ohko/chatroom/common/com"
	"github.com/ohko/chatroom/public"
	"github.com/ohko/chatroom/srv"
)

var (
	dbPath = flag.String("db", "./db/chatroom.db", "database file path,eg: postgres://postgres:postgres@localhost/chatroom?sslmode=disable&TimeZone=Asia/Shanghai")
	addr   = flag.String("s", ":8080", "server address")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Flags() | log.Lshortfile)

	if err := com.Init(*dbPath); err != nil {
		log.Println(err)
	}

	go srv.PingDeamon()
	srv.HandleWS("/ws")
	srv.HandleApiFuncs("/im/")
	srv.HandleStatic("/", public.IndexFile)

	log.Println("Server listen:", *addr)
	log.Println(http.ListenAndServe(*addr, nil))
}
