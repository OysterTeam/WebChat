package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type SignUp struct {
	NickName  *string `json:"nick_name"`
	Email     *string `json:"email"`
	Gender    *string `json:"gender"`
	genderInt int
	PwdRaw    *string `json:"pwd_raw"`
}

func (s *SignUp) checkSignUpField() error {
	if s.NickName == nil || s.Email == nil || s.Gender == nil {
		return errors.New("有参数字段为空")
	} else if s.genderInt, _ = strconv.Atoi(*s.Gender); s.genderInt < 1 || s.genderInt > 2 {
		return errors.New("gender参数范围错误")
	} else if !IsEmailValid(s.Email) {
		return errors.New("email参数不合法")
	} else if len(*s.PwdRaw) < 6 {
		return errors.New("密码过短")
	}
	return nil
}

func serveSignUp(s *ChatServer, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	SetupCORS(&w)
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
				HttpResponseData: false,
			}

		} else if err = sinUp.checkSignUpField(); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				HttpResponseMsg:  "错误: " + fmt.Sprint(err),
				HttpResponseData: false,
			}
		} else if err = s.db.CreateUser(sinUp.NickName, sinUp.Email, sinUp.PwdRaw, sinUp.genderInt); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				HttpResponseMsg:  "错误: " + fmt.Sprint(err),
				HttpResponseData: false,
			}
		} else {
			w.WriteHeader(http.StatusCreated)
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusCreated,
				HttpResponseMsg:  "成功: 已创建用户",
				HttpResponseData: true,
			}
		}
		hrj, _ := json.Marshal(hr)
		_, _ = w.Write(hrj)
	}
}

type SignIn struct {
	UserID *string `json:"user_id"`
	Email  *string `json:"email"`
	PwdRaw *string `json:"pwd_raw"`
}

func (s *SignIn) checkSignInField() error {
	if s.UserID == nil && s.Email == nil {
		return errors.New("user_id 与 email 不能同时为空")
	} else if s.PwdRaw == nil {
		return errors.New("pwd_raw不能为空")
	}
	return nil
}

func serveSignIn(s *ChatServer, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	SetupCORS(&w)
	var hr HttpResponseJson
	w.Header().Set("Content-type", "application/json")
	if r.Method == http.MethodGet {
		hr = HttpResponseJson{
			HttpResponseCode: http.StatusOK,
			HttpResponseMsg:  "请使用POST方法",
			HttpResponseData: false,
		}
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal("read request error")
		}
		var sinIn SignIn
		err = json.Unmarshal(body, &sinIn)
		if err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				HttpResponseMsg:  "错误: json错误无法解析",
				HttpResponseData: false,
			}

		} else if err = sinIn.checkSignInField(); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				HttpResponseMsg:  "错误: " + fmt.Sprint(err),
				HttpResponseData: false,
			}
		} else {
			var pwdCorrect bool
			if sinIn.Email != nil {
				pwdCorrect = s.db.VerifyPwdByEmail(sinIn.Email, sinIn.PwdRaw)
			}
			w.WriteHeader(http.StatusCreated)
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusCreated,
				HttpResponseMsg:  "密码匹配状态:" + strconv.FormatBool(pwdCorrect),
				HttpResponseData: true,
			}
		}
	}
	hrj, _ := json.Marshal(hr)
	_, _ = w.Write(hrj)
}
