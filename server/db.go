package server

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type DB interface {
	Open()
	Close()
}

type SqlDataBase struct {
	db             *sql.DB
	driverName     string
	dataSourceName string
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

func (s *SqlDataBase) ExecSql(query string) {
	res, err := s.db.Exec(query)
	if err != nil {
		log.Fatal(err, "\nerror sql: ", query)
	}
	log.Println("ExecSql Res:", res)
}

func (s *SqlDataBase) CreateUser(nickName, email *string, gender int) {
	s.ExecSql(fmt.Sprintf("INSERT INTO User(U_Nickname, U_Gender,U_Email) VALUES('%s','%d','%s')", *nickName, gender, *email))
}
