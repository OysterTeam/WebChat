package server

import (
	"database/sql"
	"errors"
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

func (s *SqlDataBase) GetUserIDByEmail(email *string) int {
	query := "SELECT `U_ID` FROM `User` WHERE `U_Email`=?"
	row := s.db.QueryRow(query, *email)
	var UserID int
	_ = row.Scan(&UserID)
	return UserID
}

func (s *SqlDataBase) CreateUser(nickName, email, pwd *string, gender int) error {
	UserID := s.GetUserIDByEmail(email)
	if UserID != 0 {
		return errors.New("email已注册")
	}
	s.createUserMutex.Lock()
	query := "INSERT INTO `User`(U_Nickname, U_Gender,U_Email) VALUES(?,?,?)"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(*nickName, gender, *email)
	if err != nil {
		return err
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	s.createUserMutex.Unlock()
	query = "INSERT INTO `UserPwd`(U_ID, U_Pwd) VALUES(?, ?)"
	stmt, err = s.db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(userID, *pwd)
	if err != nil {
		return err
	}
	return nil
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
