package server

import (
	"log"
	"net/http"
	"time"
)

type ChatServer struct {
	db            *SqlDataBase     // 登录数据库
	addr          *string          // 端口
	RWTimeout     time.Duration    // 读写超时
	msgMux        *MsgMux          // 消息路由
	online        chan *Client     // 上线通道
	offline       chan *Client     // 下线通道
	onlineMap     map[*Client]bool //在线用户表
	tokenUserMap  map[string]int   //token-UserID表
	userClientMap map[int]*Client  //UserID-Client表
}

func NewChatServer(addr *string) *ChatServer {
	return &ChatServer{
		db:            NewSqliteDB("db/my.db"),
		addr:          addr,
		RWTimeout:     10 * time.Second,
		msgMux:        NewMsgMux(),
		online:        make(chan *Client),
		offline:       make(chan *Client),
		onlineMap:     make(map[*Client]bool),
		tokenUserMap:  make(map[string]int),
		userClientMap: make(map[int]*Client),
	}
}

func (s *ChatServer) Run() {
	httpServer := &http.Server{Addr: *s.addr, Handler: nil,
		ReadTimeout: s.RWTimeout, WriteTimeout: s.RWTimeout}
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(s, w, r)
	})
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		serveSignUp(s, w, r)
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		serveSignIn(s, w, r)
	})
	s.db.Open()
	defer s.db.Close()
	go s.msgMux.Serve()
	go func() {
		for {
			select {
			case client := <-s.online:
				s.onlineMap[client] = true
				log.Println("有用户建立连接，当前在线用户总数：", len(s.onlineMap))
			case client := <-s.offline:
				if _, ok := s.onlineMap[client]; ok {
					delete(s.onlineMap, client)
					log.Println("有用户退出连接，当前在线用户总数：", len(s.onlineMap))
				}
			}
		}
	}()
	_ = httpServer.ListenAndServe()
}
