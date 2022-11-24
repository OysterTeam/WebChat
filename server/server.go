package server

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type ChatServer struct {
	db                *SqlDataBase  // 用户数据库
	addr              *string       // 端口
	RWTimeout         time.Duration // 读写超时
	msgMux            *MsgMux       // 消息路由
	signIn            chan int      // 上线通道
	signOut           chan int      // 下线通道
	onlineUserNum     int           // 在线用户数量
	userTokenSyncMap  sync.Map      // UserID-token 对应表
	userClientSyncMap sync.Map      // UserID-client 对应表
}

func NewChatServer(addr *string) *ChatServer {
	server := ChatServer{
		db:        NewSqliteDB("db/my.db"),
		addr:      addr,
		RWTimeout: 10 * time.Second,
		signIn:    make(chan int),
		signOut:   make(chan int),
	}
	server.msgMux = NewMsgMux(&server)
	return &server
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
			case uid := <-s.signIn:
				s.onlineUserNum++
				log.Println("用户", uid, "建立连接，当前在线用户总数：", s.onlineUserNum)
			case uid := <-s.signOut:
				s.onlineUserNum--
				s.userClientSyncMap.Delete(uid)
				log.Println("用户", uid, "退出连接，当前在线用户总数：", s.onlineUserNum)
			}
		}
	}()
	_ = httpServer.ListenAndServe()
}

func (s *ChatServer) VerifyToken(userID int, tokenStr *string) bool {
	realToken, ok := s.userTokenSyncMap.Load(userID)
	if !ok {
		return false
	}
	if *tokenStr == realToken {
		return true
	}
	return false
}
