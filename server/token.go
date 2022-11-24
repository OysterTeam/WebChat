package server

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

var ExpiresSecond int64 = 60

var secretSigningKey = []byte("secretSigningKey")

func GetTokenStr(userID string) *string {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + ExpiresSecond,
		Subject:   userID,
		Issuer:    "server",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secretSigningKey)
	if err != nil {
		log.Println(err)
	}
	return &tokenStr
}

func ParseTokenStr(tokenStr *string) (*string, error) {
	if *tokenStr == "" {
		return nil, errors.New("空token")
	}
	token, err := jwt.ParseWithClaims(*tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretSigningKey, nil
	})
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return &claims.Subject, nil
	}
	return nil, err
}
