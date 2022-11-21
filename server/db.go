package server

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

type DB interface {
	Open()
	Close()
}

type SqlDataBase struct {
	db              *sql.DB
	driverName      string
	dataSourceName  string
	createUserMutex sync.Mutex
}

func NewSqliteDB(path string) *SqlDataBase {
	return &SqlDataBase{
		dataSourceName: path,
		driverName:     "sqlite3",
	}
}

func (s *SqlDataBase) Open() {
	var err error
	s.db, err = sql.Open(s.driverName, s.dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SqlDataBase) Close() {
	err := s.db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SqlDataBase) ExecSql(query string) sql.Result {
	res, err := s.db.Exec(query)
	if err != nil {
		log.Println(err, "\nerror sql: ", query)
	}
	return res
}

func (s *SqlDataBase) CreateUser(nickName, email, pwd *string, gender int) { // 还需要校验email是否已经注册
	s.createUserMutex.Lock()
	res := s.ExecSql(fmt.Sprintf("INSERT INTO `User`(U_Nickname, U_Gender,U_Email) VALUES('%s','%d','%s')", *nickName, gender, *email))
	userID, _ := res.LastInsertId()
	s.createUserMutex.Unlock()
	s.ExecSql(fmt.Sprintf("INSERT INTO `UserPwd`(U_ID, U_Pwd) VALUES('%d', '%s')", userID, *pwd))
}

func (s *SqlDataBase) VerifyPwdByEmail(email, pwdA *string) bool {
	query := "SELECT `U_ID` FROM `User` WHERE `U_Email`=?"
	row := s.db.QueryRow(query, *email)
	var UserID int
	_ = row.Scan(&UserID)
	if UserID == 0 { // 用户不存在
		return false
	}
	query = "SELECT `U_Pwd` FROM `UserPwd` WHERE `U_ID`=?"
	row = s.db.QueryRow(query, UserID)
	var pwdB string
	_ = row.Scan(&pwdB)
	if *pwdA == pwdB {
		return true
	}
	return false
}
