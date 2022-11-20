package server

import (
	"log"
)

const (
	// 登录阶段 1000-1999
	loginCode     = 1000 // 登录阶段
	signUpCode    = 1001 // 注册
	signInCode    = 1002 // 登录
	forgetPwdCode = 1003 // 忘记密码
	resetPwdCode  = 1004 // 重设密码
	// 聊天消息 2000-2999
	msgCode       = 2000 // 聊天消息
	plainMsgCode  = 2001 // 纯文本消息
	richMsgCode   = 2002 // 富文本消息
	picMsgCode    = 2003 // 图片消息
	memeMsgCode   = 2004 // 表情包消息
	voiceMsgCode  = 2005 // 语音消息
	voiceChatCode = 2101 // 语音电话消息
	videoChatCode = 2102 // 视频电话消息
)

/*
举例
{
    "send_id":1,
    "msg_code":2001,
    "msg_from":30000001,
    "msg_to":30000002,
    "msg_to_group":false,
    "msg_content":"hello"
}
*/

type Msg struct {
	SendId     int    `json:"send_id"`
	MsgCode    int    `json:"msg_code"`
	MsgFrom    int    `json:"msg_from"`
	MsgTo      int    `json:"msg_to"`
	MsgToGroup bool   `json:"msg_to_group"`
	MsgContent []byte `json:"msg_content"`
}

type MsgMux struct {
	msgQueue chan *Msg
}

func NewMsgMux() *MsgMux {
	return &MsgMux{msgQueue: make(chan *Msg, 64)}
}

func (s *MsgMux) Serve() {
	for {
		select {
		case msg := <-s.msgQueue:
			log.Println("MsgMux接收到消息", msg, string(msg.MsgContent))
		}
	}
}
