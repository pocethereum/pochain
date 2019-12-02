package minedev

import (
	"time"
)

var sql_create_host string = `
	CREATE TABLE IF NOT EXISTS t_host (
	F_id INTEGER PRIMARY KEY AUTOINCREMENT,
	F_Hostname TEXT NULL,
	F_Status INTEGER NOT NULL DEFAULT 0,
	F_CreateTime TEXT NULL,
	F_ModifyTime TEXT NULL)
`

type Host struct {
	Hostname   string `json:"hostname"`
	Status     string `json:"status"`
	CreateTime string `json:"create_time"`
	ModifyTime string `json:"modify_time"`
}

func (h *Host) QueryHostInfo() (e error) {
	db := GetDbInstance()
	sql := `
		select F_Hostname, F_Status, F_CreateTime, F_ModifyTime
		from t_host WHERE F_id = 1
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return err
	}

	err = stmt.QueryRow().Scan(&h.Hostname, &h.Status, &h.CreateTime, &h.ModifyTime)
	if err != nil {
		return err
	}

	return nil
}

func (h *Host) ModHostInfo() (e error) {
	db := GetDbInstance()
	sql := `
		replace into t_host(F_id, F_Hostname, F_Status, F_CreateTime, F_ModifyTime)
		values(1,?,?,?,?)
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return err
	}
	h.ModifyTime = time.Now().Format(DefaultTimeFormat)
	_, e = stmt.Exec(h.Hostname, h.Status, h.CreateTime, h.ModifyTime)
	return
}

func (h *Host) ClsHostInfo() (e error) {
	db := GetDbInstance()
	sql := `
		delete from t_host
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return err
	}
	_, e = stmt.Exec()
	return
}
