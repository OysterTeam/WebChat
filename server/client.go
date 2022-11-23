package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	// 向client端发送数据的最大时延
	writeWait = 10 * time.Second

	// 接收client端发送的pong的最大时延
	pongWait = 60 * time.Second

	// ping client端的周期，必须比 pongWait 短。
	pingPeriod = (pongWait * 9) / 10

	// 消息大小限制。
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(*http.Request) bool { return true },
}

type Client struct {
	chatServer *ChatServer
	conn       *websocket.Conn
	msgMux     *MsgMux
	userId     int
}

func serveWs(s *ChatServer, w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization") //Bearer
	//if tokenStr == "" {                       //未传token
	//	return
	//}
	//userID, ok := s.tokenUserMap[tokenStr]
	userIDStr, err := ParseTokenStr(&tokenStr)
	if err != nil {
		return
	}
	userID, err := strconv.Atoi(*userIDStr)
	if err != nil {
		return
	}
	if ok := s.VerifyToken(userID, &tokenStr); !ok {
		return
	}
	//建立连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{ // 未赋值userId，代表仅建立ws连接，未登录用户
		chatServer: s,
		conn:       conn,
		msgMux:     s.msgMux,
		userId:     userID,
	}
	s.userClientSyncMap.Store(userID, client)
	client.chatServer.signIn <- userID
	_ = conn.WriteMessage(websocket.TextMessage, append([]byte("websocket连接成功")))
	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.chatServer.signOut <- c.userId
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var msg Msg
		if err := json.Unmarshal(message, &msg); err != nil {
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = c.conn.WriteMessage(websocket.TextMessage, append([]byte("wrong json message: "), message...))
			continue
		}

		c.msgMux.msgQueue <- &msg
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		//_ = c.conn.WriteMessage(websocket.TextMessage, append([]byte("服务器接收到数据："), message...))
		_ = c.conn.WriteMessage(websocket.TextMessage, append([]byte("receiveMsgId:"), []byte(strconv.Itoa(msg.SendId))...))
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
