package core

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type dbPool struct {
	isUsed bool
	db     *sql.DB
}

var dbPoolStorage map[int]*dbPool

//数据库连接池实例化
func DbNew(num int) {
	dbPoolStorage = make(map[int]*dbPool, num)
	for i := 0; i < num; i++ {
		db, err := sql.Open("mysql", "dbusername:db_pwd@tcp(db_servername)/db")
		if err != nil {
			//这里一定要panic终止程序运行
			panic(err.Error())
		}
		dbPoolStorage[i] = &dbPool{db: db, isUsed: false}
	}
}

//获取数据库连接实例
func GetDb() *sql.DB {
	for _, v := range dbPoolStorage {
		if v.isUsed == false {
			v.isUsed = true
			return v.db
		}
	}
	return nil
}

//数据库连接还回连接池
func FreeDb(db *sql.DB) bool {
	for _, v := range dbPoolStorage {
		if v.db == db {
			v.isUsed = false
			return true
		}
	}
	return false
}
