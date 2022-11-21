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

func (s *SqlDataBase) CreateUser(nickName, email, pwd *string, gender int) {
	s.createUserMutex.Lock()
	res := s.ExecSql(fmt.Sprintf("INSERT INTO User(U_Nickname, U_Gender,U_Email) VALUES('%s','%d','%s')", *nickName, gender, *email))
	userID, _ := res.LastInsertId()
	s.createUserMutex.Unlock()
	s.ExecSql(fmt.Sprintf("INSERT INTO UserPwd(U_ID, U_Pwd) VALUES('%d', '%s')", userID, *pwd))
}
