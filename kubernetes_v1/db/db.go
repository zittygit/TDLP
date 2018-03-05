package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var MysqlDB *sql.DB

func InitMysqlDB(dbstr string) {
	var err error
	MysqlDB, err = sql.Open("mysql", dbstr)
	if err != nil {
		panic(err)
	}
	err = MysqlDB.Ping()
	if err != nil {
		panic(err)
	}
}

func CloseMysqlDB() {
	err := MysqlDB.Close()
	if err != nil {
		panic(err)
	}
}
