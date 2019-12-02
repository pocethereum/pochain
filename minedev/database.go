package minedev

import (
	"database/sql"
	"github.com/pocethereum/pochain/log"
	_ "github.com/mattn/go-sqlite3"
	"path"
	"strings"
	"sync"
)

const (
	DefaultTimeFormat = "2006-01-02 15:04:05"
)
const (
	STATUS_OK      = 0
	STATUS_DELETED = 1
	STATUS_INVALID = 99
)

var (
	DefaultDescription = "description.xml"
	DefaultDatabase    = "database.sqlite"
	RWLock             sync.RWMutex
	instance           *sql.DB
)

func GetDbInstance() *sql.DB {
	if instance == nil {
		var err error
		instance, err = sql.Open("sqlite3", DefaultDatabase)
		if err != nil {
			log.Error("open db file error", "err", err.Error())
			return nil
		}
		log.Info("init sqlite db done", "database", DefaultDatabase)
	}
	return instance
}

func Init(datadir string) {
	if !strings.HasPrefix(DefaultDescription, datadir) {
		DefaultDescription = path.Join(datadir, DefaultDescription)
	}

	if !strings.HasPrefix(DefaultDatabase, datadir) {
		DefaultDatabase = path.Join(datadir, DefaultDatabase)
	}

	if _, err := execSql(sql_create_user); err != nil {
		panic("execSql(sql_create_user) Failed:" + err.Error())
	}

	if _, err := execSql(sql_create_host); err != nil {
		panic("execSql(sql_create_host) Failed:" + err.Error())
	}

	if _, err := execSql(sql_create_plot); err != nil {
		panic("execSql(sql_create_plot) Failed:" + err.Error())
	}
}

func execSql(sql string) (rowAffect int64, err error) {
	RWLock.Lock()
	defer RWLock.Unlock()
	db := GetDbInstance()
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		log.Error("execSql", "prepare sql", sql, "err", err.Error())
		return 0, err
	}
	res, err := stmt.Exec()
	if err != nil {
		log.Error("execSql", "exec sql", sql, "err", err.Error())
		return 0, err
	}
	return res.RowsAffected()
}

func CloseStmt(s *sql.Stmt) {
	if s != nil {
		s.Close()
	}
}

func CloseRows(r *sql.Rows) {
	if r != nil {
		r.Close()
	}
}
