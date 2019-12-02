package minedev

import (
	"encoding/json"
	"time"
)

type Uatk struct {
	SessionId     string `json:"sessionid"`
	SessionValues string `json:"sessionvalues"`
	CreateTime    string `json:"create_time"`
	ModifyTime    string `json:"modify_time"`
}

type UatkInfo struct {
	Username string `json:"username"`
}

func (u *Uatk) QuerySessionValues(uatk string) (uatkinfo UatkInfo, e error) {
	db := GetDbInstance()
	sql := `
		select F_SessionId, F_SessionValues, F_CreateTime, F_ModifyTime
		from t_uatk where SessionId = ?
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return uatkinfo, err
	}

	err = stmt.QueryRow(uatk).Scan(&u.SessionId, &u.SessionValues, &u.CreateTime, &u.ModifyTime)
	if err != nil {
		return uatkinfo, err
	}

	err = json.Unmarshal([]byte(u.SessionValues), &uatkinfo)
	if err != nil {
		return uatkinfo, err
	}

	return uatkinfo, nil
}

func (u *Uatk) InsertSessionValues() (e error) {
	db := GetDbInstance()
	sql := `
		insert into t_uatk(F_SessionId, F_SessionValues, F_CreateTime, F_ModifyTime)
		values(?,?,?,?)
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return err
	}
	u.CreateTime = time.Now().Format(DefaultTimeFormat)
	u.ModifyTime = time.Now().Format(DefaultTimeFormat)
	_, e = stmt.Exec(u.SessionId, u.SessionValues, u.CreateTime, u.ModifyTime)
	return
}
