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
			BoolStatus:       false,
			ResponseMsg:      "请使用POST方法",
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
				BoolStatus:       false,
				ResponseMsg:      "错误: json错误无法解析",
			}

		} else if err = sinUp.checkSignUpField(); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				BoolStatus:       false,
				ResponseMsg:      "错误: " + fmt.Sprint(err),
			}
		} else if uid, err := s.db.CreateUser(sinUp.NickName, sinUp.Email, sinUp.PwdRaw, sinUp.genderInt); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				BoolStatus:       false,
				ResponseMsg:      "错误: " + fmt.Sprint(err),
			}
		} else {
			w.WriteHeader(http.StatusCreated)
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusCreated,
				BoolStatus:       true,
				ResponseMsg:      "成功: 已创建用户，UserID请见ResponseData字段。",
				ResponseData:     uid,
			}
		}
		hrj, _ := json.Marshal(hr)
		_, _ = w.Write(hrj)
	}
}

type SignIn struct {
	UserID    *string `json:"user_id"`
	UserIDInt int
	Email     *string `json:"email"`
	PwdRaw    *string `json:"pwd_raw"`
}

func (s *SignIn) checkSignInField() error {
	if s.UserID == nil && s.Email == nil {
		return errors.New("user_id 与 email 不能同时为空")
	} else if s.UserID != nil && s.Email != nil {
		return errors.New("user_id 与 email 不能同时填写")
	} else if s.PwdRaw == nil {
		return errors.New("pwd_raw不能为空")
	}
	var err error
	if s.UserID != nil {
		s.UserIDInt, err = strconv.Atoi(*s.UserID)
	}
	return err
}

func serveSignIn(s *ChatServer, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	SetupCORS(&w)
	var hr HttpResponseJson
	w.Header().Set("Content-type", "application/json")
	if r.Method == http.MethodGet {
		hr = HttpResponseJson{
			HttpResponseCode: http.StatusOK,
			BoolStatus:       false,
			ResponseMsg:      "请使用POST方法",
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
				BoolStatus:       false,
				ResponseMsg:      "错误: json错误无法解析",
			}

		} else if err = sinIn.checkSignInField(); err != nil {
			hr = HttpResponseJson{
				HttpResponseCode: http.StatusBadRequest,
				BoolStatus:       false,
				ResponseMsg:      "错误: " + fmt.Sprint(err),
			}
		} else { // json校验通过，校验密码
			var pwdCorrect bool
			var info string
			var userIDInt int
			if sinIn.UserIDInt != 0 {
				pwdCorrect, info = s.db.VerifyPwdByUserID(sinIn.UserIDInt, sinIn.PwdRaw)
				userIDInt = sinIn.UserIDInt
			} else {
				pwdCorrect, info, userIDInt = s.db.VerifyPwdByEmail(sinIn.Email, sinIn.PwdRaw)
			}
			if !pwdCorrect { // 密码不正确
				w.WriteHeader(http.StatusOK)
				hr = HttpResponseJson{
					HttpResponseCode: http.StatusOK,
					BoolStatus:       pwdCorrect,
					ResponseMsg:      "密码匹配状态:" + info,
				}
			} else { // 密码正确，判断是否已有登录
				_, ok := s.userClientSyncMap.Load(userIDInt)
				if ok { //存在连接，目前不允许顶替登录
					w.WriteHeader(http.StatusCreated)
					hr = HttpResponseJson{
						HttpResponseCode: http.StatusOK,
						BoolStatus:       false,
						ResponseMsg:      "密码匹配，但已有连接。",
					}
				} else { // 可以连接，发送token
					tokenStr := GetTokenStr(strconv.Itoa(userIDInt))
					w.WriteHeader(http.StatusCreated)
					hr = HttpResponseJson{
						HttpResponseCode: http.StatusCreated,
						ResponseMsg:      "密码匹配状态:" + info + "。若匹配，ResponseData字段为token。",
						BoolStatus:       pwdCorrect,
						ResponseData:     *tokenStr,
					}
					s.userTokenSyncMap.Store(userIDInt, *tokenStr)
				}
			}
		}
	}
	hrj, _ := json.Marshal(hr)
	_, _ = w.Write(hrj)
}
