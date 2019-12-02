package minedev

import (
	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/consensus"
	"github.com/pocethereum/pochain/consensus/poc"
	"github.com/pocethereum/pochain/consensus/poc/plotter"
	"github.com/pocethereum/pochain/eth/downloader"
	"github.com/pocethereum/pochain/log"
	"github.com/pocethereum/pochain/miner"
	"encoding/json"
	"os"
	"time"
)

type MineDevice struct {
	ieth              IEthereum
}

type IEthereum interface {
	StopMining()
	IsMining() bool
	Miner() *miner.Miner
	IsPloting() bool
	Downloader() *downloader.Downloader
	Plotter() *plotter.Plotter
	Engine() consensus.Engine
}

type Result map[string]interface{}

const (
	DEV_STATUS_UNBIND   = "unbind"
	DEV_STATUS_BINDED   = "binded"
	DEV_STATUS_SYNCING  = "syncing"
	DEV_STATUS_PLOTTING = "plotting"
	DEV_STATUS_MINNING  = "minning"
	DEV_STATUS_WAITING  = "waiting"
)

var PlotpathsUpdater = make(chan string, 100)
var gDev *MineDevice

func New(ieth IEthereum) *MineDevice {
	gDev = &MineDevice{ieth: ieth}
	PlotpathsUpdater<-GetSettingPlotdirs()
	return gDev
}

func (dev *MineDevice) status() string {
	// Case 1.
	uo := User{}
	binduser, _ := uo.QueryBindUser()
	if binduser.Username == "" {
		return DEV_STATUS_UNBIND
	}

	// Case 2.
	isplotting := dev.ieth.IsPloting()
	host := Host{}
	host.QueryHostInfo()
	if host.Hostname == "" && !isplotting {
		return DEV_STATUS_BINDED
	}

	// Case 3.
	if isplotting {
		return DEV_STATUS_PLOTTING
	} else if dev.getPlotedSize() == 0 {
		return DEV_STATUS_BINDED
	}

	// Case 4.
	progress := dev.ieth.Downloader().Progress()
	if progress.CurrentBlock < progress.HighestBlock {
		return DEV_STATUS_SYNCING
	}

	// Case 5.
	if dev.ieth.IsMining() {
		return DEV_STATUS_MINNING
	}

	return DEV_STATUS_WAITING
}

func (dev *MineDevice) getPlotedSize() uint64 {
	engine := dev.ieth.Engine()
	if pocengine, ok := engine.(*poc.Poc); ok {
		return pocengine.GetSize()
	}
	return 0
}

func accessable(uatk string, username string) (uinfo UatkInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	u := Uatk{}
	uinfo, err = u.QuerySessionValues(uatk)
	if err != nil {
		log.Info("Accessable failed, QuerySessionValues error.", "error", err.Error())
		return uinfo, err
	}

	return uinfo, nil
}

func (dev *MineDevice) Status() (r Result) {
	r = Result{"err": "ok"}
	h := Host{}
	h.QueryHostInfo()
	u := User{}
	binduser, _ := u.QueryBindUser()

	r["status"] = dev.status()
	r["miner"] = binduser.Username
	r["hostname"] = h.Hostname
	r["plotsize"] = dev.getPlotedSize()

	if r["status"] == DEV_STATUS_PLOTTING {
		r["progress"] = dev.ieth.Plotter().Progress()
	}
	return r
}

func (dev *MineDevice) Bind(miner common.Address, auth string) (r Result) {
	r = Result{"err": "ok"}

	u := User{}
	u.Username = miner.String()
	u.Password = auth

	uo := User{}
	binduser, _ := uo.QueryBindUser()
	if binduser.Username != "" {
		r["err"] = "Have been binded by " + getEncodedBindUserName(binduser.Username)
		return r
	}

	u.CreateTime = time.Now().Format(DefaultTimeFormat)
	if err := u.ModBindUser(); err != nil {
		log.Info("Bind User Save Failed", "error", err.Error())
		r["err"] = "Bind User Save Failed:" + err.Error()
		return
	}

	GetInstance().RefreshDeviceData()
	return r
}

func (dev *MineDevice) Unbind(miner common.Address, auth string) (r Result) {
	r = Result{"err": "ok"}
	u := User{}
	u.Username = miner.String()

	uo := User{}
	binduser, _ := uo.QueryBindUser()
	if binduser.Username == "" {
		r["err"] = "Cannot unbind by:" + getEncodedBindUserName(u.Username) + ", not binded yet"
		return r
	} else if binduser.Username != u.Username {
		r["err"] = "Cannot unbind by:" + getEncodedBindUserName(u.Username) + ", have binded by:" + getEncodedBindUserName(binduser.Username)
		return r
	}

	if err := u.ClsUser(); err != nil {
		log.Info("Unbind User Save Failed", "error", err.Error())
		r["err"] = "Unbind User Save Failed:" + err.Error()
		return
	}

	GetInstance().RefreshDeviceData()
	return r
}

func (dev *MineDevice) Restart() (r Result) {
	r = Result{"err": "ok"}
	process, _ := os.FindProcess(os.Getpid())
	process.Kill()
	return r
}

type InputSettingAction string

const (
	ACTION_SELECT   = "SELECT"
	ACTION_UNSELECT = "UNSELECT"
	ACTION_RMDATA   = "RMDATA"
)

type InputSettingPlotdir struct {
	Plot
	Action InputSettingAction `json:"action"`
}

func (dev *MineDevice) Setting(name string, value string) (r Result) {
	r = Result{"err": "ok"}
	log.Info("Setting", "name", name, "value", value)

	switch name {
	case "HOSTNAME":
		r = dev.SettingHostname(value)
	case "PLOTDIRS":
		r = dev.SettingPlotdirs(value)
	}

	GetInstance().RefreshDeviceData()
	return r
}

func (dev *MineDevice) Getting(name string) (r Result) {
	r = Result{"err": "ok"}
	log.Info("Getting", "name", name)

	switch name {
	case "HOSTNAME":
		r = dev.GettingHostname()
	case "PLOTDIRS":
		r = dev.GettingPlotdirs()
	}

	return r
}

func (dev *MineDevice) GettingHostname() (r Result) {
	r = Result{"err": "ok"}

	h := Host{}
	h.QueryHostInfo()
	r["value"] = h.Hostname
	return r
}

func (dev *MineDevice) SettingHostname(hostname string) (r Result) {
	r = Result{"err": "ok"}

	h := Host{}
	h.QueryHostInfo()
	h.Hostname = hostname

	if err := h.ModHostInfo(); err != nil {
		log.Info("Setting Host Save Failed", "error", err.Error())
		r["err"] = "Setting Host Save Failed:" + err.Error()
		return
	}

	return r
}

func (dev *MineDevice) GettingPlotdirs() (r Result) {
	r = Result{"err": "ok"}
	u := User{}
	binduser, _ := u.QueryBindUser()

	//查询数据库中的记录
	dbplots, err := (&Plot{PlotSeed: binduser.Username}).QueryAllPlotInfo()
	if err != nil {
		log.Info("QueryAllPlotInfo Failed", "error", err.Error())
		r["err"] = "QueryAllPlotInfo Failed:" + err.Error()
		return
	}

	//查询实际挂载的目录
	avplots, err := getAvaliablePlotDiskInfo()
	if err != nil {
		log.Info("getAvaliablePlotDiskInfo Failed", "error", err.Error())
		r["err"] = "getAvaliablePlotDiskInfo Failed:" + err.Error()
		return
	}

	retplotdirs := []Plot{}
	for _, avp := range avplots {
		rowp := avp
		log.Info("GettingPlotdirs=============", "rowp", rowp)
		for _, dbp := range dbplots {
			log.Info("GettingPlotdirs=============", "dbp", dbp)
			if avp.Path == dbp.Path {
				rowp.Id = dbp.Id
				rowp.PlotSize = dbp.PlotSize
				rowp.Status = dbp.Status
				break
			}
		}
		retplotdirs = append(retplotdirs, rowp)
	}
	//value, err := json.Marshal(retplotdirs)
	//if err != nil {
	//	log.Info("Marshal Failed", "error", err.Error())
	//	r["err"] = "Marshal Failed:" + err.Error()
	//	return
	//}

	r["value"] = retplotdirs //string(value)
	return r
}

func (dev *MineDevice) SettingPlotdirs(value string) (r Result) {
	r = Result{"err": "ok"}

	retmap := dev.Status()
	if retmap["err"] != "ok" || retmap["miner"] == "" {
		log.Info("SettingPlotdirs Failed, have not binded yet")
		r["err"] = "have not binded yet"
		return
	}

	plotsettings := []InputSettingPlotdir{}
	if err := json.Unmarshal([]byte(value), &plotsettings); err != nil {
		log.Info("Unmarshal Failed", "error", err.Error(), "value", value)
		r["err"] = "SettingPlotdirs Failed:" + err.Error()
		return
	}

	for _, setting := range plotsettings {
		rowp := Plot{
			Id:       setting.Id,
			Name:     setting.Name,
			Path:     setting.Path,
			Uuid:     setting.Uuid,
			PlotSize: setting.PlotSize,
			Status:   setting.Status,
		}

		/////// TODO: change input from frontend or generate backend ////begin//
		rowp.PlotSize = 8 * 1024 * 1024 * 1024 //test for 8GB
		rowp.PlotSeed = retmap["miner"].(string)
		/////////////////////////////////////////////////////////////////end////

		switch setting.Action {
		case ACTION_SELECT:
			rowp.PlotSelectedByIds(rowp.Id, rowp.Path)
			dev.ieth.Plotter().Reload()
		case ACTION_UNSELECT:
			rowp.PlotUnselectedByIds(rowp.Id, rowp.Path)
			dev.ieth.Plotter().Reload()
		case ACTION_RMDATA:
		}
	}

	PlotpathsUpdater<-GetSettingPlotdirs()
	return r
}

func (dev *MineDevice) Login() (r Result) {
	return
}
