package minedev

import (
	"database/sql"
	"github.com/pocethereum/pochain/consensus/poc/plotter"
	"github.com/pocethereum/pochain/log"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

var sql_create_plot string = `
	CREATE TABLE IF NOT EXISTS t_plot (
	F_Id INTEGER PRIMARY KEY AUTOINCREMENT,
	F_Name TEXT NOT NULL DEFAULT "",
	F_Path TEXT NOT NULL DEFAULT "",
	F_Uuid TEXT NOT NULL DEFAULT "",
	F_PlotSeed TEXT NOT NULL DEFAULT "",
	F_PlotDir TEXT NOT NULL DEFAULT "",
	F_PlotSize TEXT NOT NULL DEFAULT "",
	F_PlotParam TEXT NOT NULL DEFAULT "",
	F_Status INTEGER NOT NULL DEFAULT 0,
	F_CreateTime TEXT NULL,
	F_ModifyTime TEXT NULL)
`

type Plot struct {
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Uuid      string `json:"uuid"`
	DiskSize  uint64 `json:"disksize"`
	FreeSize  uint64 `json:"freesize"`
	PlotSeed  string `json:"plotseed"`
	PlotSize  uint64 `json:"plotsize"`
	PlotDir   string `json:"plotdir"`
	PlotParam string `json:"plotparam"`
	Status    int    `json:"status"`
}

const (
	PLOT_STATUS_UNUSED   = 0 //未选中，未使用
	PLOT_STATUS_PLOTTING = 1 //已选中，p盘中
	PLOT_STATUS_PAUSED   = 2 //已选中，p盘暂停中
	PLOT_STATUS_STOPED   = 3 //未选中，p盘主动取消
	PLOT_STATUS_DONE     = 9 //已选中，p盘完成
)
const (
	DEFAULT_PLOTDIR = "${MOUNTPOINT}/plotdata/"
)

func (p *Plot) QueryAllPlotInfo() (plots []Plot, e error) {
	db := GetDbInstance()
	sql := `
		select F_Id, F_Name, F_Path, F_Uuid, F_PlotSeed, F_PlotDir, F_PlotSize, F_PlotParam, F_Status
		from t_plot where (1 = ? or F_PlotSeed = ?) COLLATE NOCASE
	`
	stmt, err := db.Prepare(sql)
	defer CloseStmt(stmt)
	if err != nil {
		log.Info("QueryAllPlotInfo error", "error", err.Error())
		return plots, err
	}
	queryall := 0
	if p.PlotSeed == "*" {
		queryall = 1
	} else {
		queryall = 0
	}

	if rows, err := stmt.Query(queryall, p.PlotSeed); err != nil {
		log.Info("QueryAllPlotInfo error", "err", err.Error(), "sql", sql)
		return plots, err
	} else {
		for rows.Next() {
			rowp := Plot{}
			err = rows.Scan(
				&rowp.Id,
				&rowp.Name,
				&rowp.Path,
				&rowp.Uuid,
				&rowp.PlotSeed,
				&rowp.PlotDir,
				&rowp.PlotSize,
				&rowp.PlotParam,
				&rowp.Status)
			if err != nil {
				log.Info("QueryAllPlotInfo error", "err", err.Error(), "sql", sql)
				return plots, err
			}
			plots = append(plots, rowp)
		}
	}

	return plots, nil
}

func (p *Plot) InsertNew() (err error) {
	log.Info("start InsertNew", "path", p.Path)

	db := GetDbInstance()
	sqlstr := `
		INSERT INTO t_plot(F_Name, F_Path, F_Uuid, F_PlotSeed, F_PlotDir, F_PlotSize, F_Status, F_CreateTime, F_ModifyTime)
		VALUES(              ?,       ?,      ?,     ?,            ?,             ?,        ?,           ?,           ?)
	`
	stmt, err := db.Prepare(sqlstr)
	defer CloseStmt(stmt)
	if err != nil {
		log.Info("InsertNew error", "error", err.Error())
		return err
	}
	var plotdir = strings.Replace(DEFAULT_PLOTDIR, "${MOUNTPOINT}", "", 1)
	var createtime = time.Now().Format("2006-01-02 15:04:05")
	if result, e := stmt.Exec(p.Name, p.Path, p.Uuid, p.PlotSeed, plotdir, p.PlotSize, p.Status, createtime, createtime); e != nil {
		log.Info("Exec sql error", "err", err.Error(), "sql", sqlstr)
		err = e
	} else {
		id := int64(0)
		id, err = result.LastInsertId()
		p.Id = uint64(id)
		log.Info("Exec insert sql success", "LastInsertId", id)
	}
	return
}

func (p *Plot) QueryByIds(id uint64, path string) (err error) {
	db := GetDbInstance()

	sqlstr := `
		select F_Id, F_Name, F_Path, F_Uuid, F_PlotSeed, F_PlotDir, F_PlotSize, F_Status
		from t_plot where (F_Id = ? or F_Path = ?) AND (1 = ? or F_PlotSeed = ?) COLLATE NOCASE
	`
	stmt, err := db.Prepare(sqlstr)
	defer CloseStmt(stmt)
	if err != nil {
		log.Info("QueryByIds error", "error", err.Error())
		return err
	}
	queryall := 0
	if p.PlotSeed == "*" {
		queryall = 1
	} else {
		queryall = 0
	}
	if err = stmt.QueryRow(id, path, queryall, p.PlotSeed).Scan(&p.Id, &p.Name, &p.Path, &p.Uuid, &p.PlotSeed, &p.PlotDir, &p.PlotSize, &p.Status); err == sql.ErrNoRows {
		log.Info("Can not find plot records", "id", id, "path", path, "sql", sqlstr)
	} else if err != nil {
		log.Info("Exec sql error", "err", err.Error(), "sql", sqlstr)
	}
	return err
}

func (p *Plot) UpdateByIds(id uint64) (err error) {
	db := GetDbInstance()
	sqlstr := `
		UPDATE t_plot Set F_Name=?, F_Path = ?, F_Uuid =?, F_PlotSeed = ?, F_PlotSize=?, F_Status = ?, F_ModifyTime = ? 
		WHERE F_Id = ?
	`
	stmt, err := db.Prepare(sqlstr)
	defer CloseStmt(stmt)
	if err != nil {
		log.Info("UpdateByIds error", "error", err.Error())
		return err
	}
	createtime := time.Now().Format("2006-01-02 15:04:05")
	if _, err = stmt.Exec(p.Name, p.Path, p.Uuid, p.PlotSeed, p.PlotSize, p.Status, createtime, id); err != nil {
		log.Info("Exec sql error", "err", err.Error(), "sql", sqlstr)
	}
	return err
}

func (plot *Plot) GetFullPlotPath() (plotpath string) {
	if plot.PlotDir == "" {
		plotpath = strings.Replace(DEFAULT_PLOTDIR, "${MOUNTPOINT}", plot.Path, 1)
	} else if strings.HasPrefix(plot.PlotDir, plot.Path) {
		plotpath = plot.PlotDir
	} else {
		plotpath = plot.Path + plot.PlotDir
	}
	return
}

func (p *Plot) PlotSelectedByIds(id uint64, path string) (err error) {
	// find record
	queryplot := *p
	if err = (&queryplot).QueryByIds(id, path); err == sql.ErrNoRows {
		err = p.InsertNew()
	}
	if err != nil {
		return err
	}
	if p.Id == 0 && queryplot.Id != 0 {
		p.Id = queryplot.Id
	}

	//Have been SELECTED
	if queryplot.Status == PLOT_STATUS_PLOTTING ||
		queryplot.Status == PLOT_STATUS_PAUSED ||
		queryplot.Status == PLOT_STATUS_DONE {
		log.Info("PlotSelectedByIds Have been SELECTED", "status", queryplot.Status)
		return
	}

	//Set to right status
	if queryplot.Status == PLOT_STATUS_UNUSED {
		p.Status = PLOT_STATUS_PLOTTING
	} else if queryplot.Status == PLOT_STATUS_STOPED {
		p.Status = PLOT_STATUS_PLOTTING
	}

	// update data
	err = p.UpdateByIds(p.Id)
	return err
}

func (p *Plot) PlotUnselectedByIds(id uint64, path string) (err error) {
	// find record
	queryplot := *p
	if err = (&queryplot).QueryByIds(id, path); err == sql.ErrNoRows {
		err = p.InsertNew()
	}
	if err != nil {
		return err
	}
	if p.Id == 0 && queryplot.Id != 0 {
		p.Id = queryplot.Id
	}

	//Have been SELECTED
	if queryplot.Status == PLOT_STATUS_UNUSED ||
		queryplot.Status == PLOT_STATUS_STOPED {
		log.Info("PlotSelectedByIds Have not SELECTED yet", "status", queryplot.Status)
		return
	}

	//Set to right status
	if queryplot.Status == PLOT_STATUS_PLOTTING {
		p.Status = PLOT_STATUS_STOPED
	} else if queryplot.Status == PLOT_STATUS_PAUSED {
		p.Status = PLOT_STATUS_STOPED
	} else if queryplot.Status == PLOT_STATUS_DONE {
		p.Status = PLOT_STATUS_UNUSED
	}

	// update data
	err = p.UpdateByIds(p.Id)
	return err
}

func (p *Plot) GetAllPlotWorks() (retworks []plotter.Work) {
	allplot := *p
	allplot.PlotSeed = "*"
	if plots, err := allplot.QueryAllPlotInfo(); err != nil {
		log.Info("GetAllPlotWorks failed", "error", err.Error())
	} else {
		for _, dbp := range plots {
			if dbp.Status != PLOT_STATUS_UNUSED && dbp.Status != PLOT_STATUS_STOPED {
				work := plotter.Work{
					Id:       strconv.FormatInt(int64(dbp.Id), 10),
					PlotDir:  dbp.GetFullPlotPath(),
					PlotSize: dbp.PlotSize,
					PlotSeed: dbp.PlotSeed,
				}
				retworks = append(retworks, work)
			}
		}
	}
	return
}

type allocRecord struct {
	startNonce uint64
	plotSize   uint64
}
type allocRecordSlice []allocRecord

func (s allocRecordSlice) Len() int           { return len(s) }
func (s allocRecordSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s allocRecordSlice) Less(i, j int) bool { return s[i].startNonce < s[j].startNonce }

type allocPlotParam struct {
	StartNonce uint64 `json:"startNonce"`
}

func (p *Plot) UpdatePlotParamById(id uint64) (err error) {
	db := GetDbInstance()
	sqlstr := `
		UPDATE t_plot Set F_PlotSize=?, F_PlotParam = ?, F_ModifyTime = ? 
		WHERE F_Id = ?
	`
	stmt, err := db.Prepare(sqlstr)
	defer CloseStmt(stmt)
	if err != nil {
		log.Info("UpdatePlotParamById error", "error", err.Error())
		return err
	}
	modifytime := time.Now().Format("2006-01-02 15:04:05")
	if _, err = stmt.Exec(p.PlotSize, p.PlotParam, modifytime, id); err != nil {
		log.Info("Exec sql error", "err", err.Error(), "sql", sqlstr)
	}
	return err
}

func (p *Plot) NonceAllocate(id, plotSeed string, plotsize uint64) (s uint64, n uint64, e error) {
	if len(plotSeed) < 5 {
		e = fmt.Errorf("plotSeed '%s' is invalid", plotSeed)
		return
	} else {
		p.PlotSeed = plotSeed
	}
	plotParam := allocPlotParam{}
	allocArray := allocRecordSlice{}

	if plots, err := p.QueryAllPlotInfo(); err != nil {
		log.Info("GetAllPlotWorks failed", "error", err.Error())
		return 0, 0, err
	} else {
		for _, p := range plots {
			if uerr := json.Unmarshal([]byte(p.PlotParam), &plotParam); uerr != nil {
				continue
			}
			// Already allocated
			if id == strconv.FormatUint(p.Id, 10) && plotsize <= p.PlotSize {
				s = plotParam.StartNonce
				n = plotsize >> 18
				e = nil
				goto allocate
			}
			allocArray = append(allocArray, allocRecord{plotParam.StartNonce, p.PlotSize})
			sort.Sort(allocArray)
		}

		// Allocate at end
		if len(allocArray) >= 1 {
			maxalloc := allocArray[len(allocArray)-1]
			s = maxalloc.startNonce + (maxalloc.plotSize >> 18)
			n = plotsize >> 18
			e = nil
		}
	}
allocate:
	plotParam.StartNonce = s
	var (
		plotParamStr, err1 = json.Marshal(plotParam)
		plotParamId, err2  = strconv.ParseInt(id, 10, 64)
	)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("Id error or Inner error")
	}
	plotForUpdate := &Plot{
		PlotSize:  plotsize,
		PlotParam: string(plotParamStr),
	}
	plotForUpdate.UpdatePlotParamById(uint64(plotParamId))

	return s, n, nil
}
