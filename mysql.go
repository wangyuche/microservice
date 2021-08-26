package main

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type SqlStruct struct {
	DB   *sql.DB
	Once sync.Once
	Info SqlInfo
}

type SqlInfo struct {
	Info         string
	MaxOpenConns int
	MaxIdleConns int
}

var writeSql = make(map[string]*SqlStruct)
var readSql = make(map[string]*SqlStruct)

func SetWriteConnectionInfo(db string, info string, maxopenconns int, maxidleconns int) {
	i := SqlInfo{Info: info, MaxOpenConns: maxopenconns, MaxIdleConns: maxidleconns}
	s := &SqlStruct{Info: i}
	writeSql[db] = s
}

func SetReadConnectionInfo(db string, info string, maxopenconns int, maxidleconns int) {
	i := SqlInfo{Info: info, MaxOpenConns: maxopenconns, MaxIdleConns: maxidleconns}
	s := &SqlStruct{Info: i}
	readSql[db] = s
}

func GetWriteConnection(db string) *sql.DB {
	_sql := writeSql[db]
	if _sql == nil {
		return nil
	}
	return getDB(_sql)
}

func GetReadConnection(db string) *sql.DB {
	_sql := readSql[db]
	if _sql == nil {
		return nil
	}
	return getDB(_sql)
}

func getDB(_sql *SqlStruct) *sql.DB {
	_sql.Once.Do(func() {
		db, err := sql.Open("mysql", _sql.Info.Info)
		if err != nil {
			panic(err.Error())
			return
		}
		err = db.Ping()
		if err != nil {
			panic(err.Error())
			db.Close()
			return
		}
		db.SetMaxOpenConns(_sql.Info.MaxOpenConns)
		db.SetMaxIdleConns(_sql.Info.MaxIdleConns)
		db.SetConnMaxLifetime(0)
		_sql.DB = db
	})
	return _sql.DB
}
