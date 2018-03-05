package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var MysqlDB *sql.DB

func InitMysqlDB(dbStr string) error {
	var err error
	MysqlDB, err = sql.Open("mysql", dbStr)
	if err != nil {
		return err
	}
	return MysqlDB.Ping()
}

func CloseMysqlDB() error {
	return MysqlDB.Close()
}
