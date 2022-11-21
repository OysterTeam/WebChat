package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type SignUp struct {
	NickName *string `json:"nick_name"`
	Email    *string `json:"email"`
	Gender   int     `json:"gender"`
}

func (s *SignUp) checkSignUpField() error {
	if s.NickName == nil || s.Email == nil || s.Gender == 0 {
		return errors.New("有参数字段为空")
	} else if s.Gender < 1 || s.Gender > 2 {
		return errors.New("gender参数范围错误")
	} else if !IsEmailValid(s.Email) {
		return errors.New("email参数不合法")
	}
	return nil
}

func serveSignUp(s *ChatServer, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var hr HttpResponseJson
	w.Header().Set("Content-type", "application/json")
	if r.Method == http.MethodGet {
		hr = HttpResponseJson{
			HttpResponseCode: http.StatusOK,
			HttpResponseMsg:  "请使用POST方法",
		}
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal("read request error")
		}
		var sinUp SignUp
		err = json.Unmarshal(body, &sinUp)
		if err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				HttpResponseMsg:  "错误: json错误无法解析",
			}

		} else if err = sinUp.checkSignUpField(); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				HttpResponseMsg:  "错误: " + fmt.Sprint(err),
			}
		} else {
			s.db.CreateUser(sinUp.NickName, sinUp.Email, sinUp.Gender)
			w.WriteHeader(http.StatusCreated)
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusCreated,
				HttpResponseMsg:  "成功: 已创建用户",
			}
		}
	}
	hrj, _ := json.Marshal(hr)
	_, _ = w.Write(hrj)
}