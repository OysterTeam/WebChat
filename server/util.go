package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"net/http"
	"regexp"
	"strings"
)

type HttpResponseJson struct {
	HttpResponseCode int         `json:"http_response_code,omitempty"`
	BoolStatus       bool        `json:"bool_status"`
	ResponseMsg      string      `json:"response_msg,omitempty"`
	ResponseData     interface{} `json:"response_data,omitempty"`
	//WSToken          string      `json:"ws_token,omitempty"`
}

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// IsEmailValid 检查邮件地址是否合法
func IsEmailValid(e *string) bool {
	if len(*e) < 3 && len(*e) > 254 {
		return false
	}
	if !emailRegex.MatchString(*e) {
		return false
	}
	parts := strings.Split(*e, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}

// SetupCORS 设置允许跨域
func SetupCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
