package minedev

var sql_create_user string = `
	CREATE TABLE IF NOT EXISTS t_user (
	F_id INTEGER PRIMARY KEY AUTOINCREMENT,
	F_Username TEXT NULL,
	F_Password TEXT NULL,
	F_CreateTime TEXT NULL,
	F_ModifyTime TEXT NULL)
`

type User struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	CreateTime string `json:"create_time"`
	ModifyTime string `json:"modify_time"`
}

func (uo *User) QueryBindUser() (u User, e error) {
	db := GetDbInstance()
	sql := `
		select F_Username, F_Password, F_CreateTime, F_ModifyTime
		from t_user
		where F_id = 1
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return u, err
	}

	_ = stmt.QueryRow().Scan(&u.Username, &u.Password, &u.CreateTime, &u.ModifyTime)
	return
}

func (uo *User) ModBindUser() (e error) {
	return uo.modBindUserImp()
}
func (uo *User) modBindUserImp() (e error) {
	db := GetDbInstance()
	sql := `
		replace into t_user(F_id, F_Username, F_Password, F_CreateTime, F_ModifyTime)
		values(1,?,?,?,?)
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return err
	}
	_, e = stmt.Exec(uo.Username, uo.Password, uo.CreateTime, uo.ModifyTime)
	return
}

func (uo *User) ClsUser() (e error) {
	return uo.clsUserImp()
}

func (uo *User) clsUserImp() (e error) {
	db := GetDbInstance()
	sql := `
		delete from t_user
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		return err
	}
	_, e = stmt.Exec()
	return
}
