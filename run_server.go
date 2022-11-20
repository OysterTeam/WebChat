package main

import (
	"WebChat/server"
	"flag"
)

var addr = flag.String("addr", ":7001", "http service address")

func main() {
	s := server.NewChatServer(addr)
	s.Run()
}
