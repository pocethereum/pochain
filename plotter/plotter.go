package plotter

import (
	"sync/atomic"

	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/log"
	"io/ioutil"
	"os"
	"strings"
)

type Plotter struct {
	worker      *worker
	seed        common.Address
	ploting     int32
	plotdataDir string
}

func New(addr common.Address, dir string) *Plotter {
	plotter := &Plotter{
		worker:      newWorker(),
		seed:        addr,
		ploting:     0,
		plotdataDir: dir,
	}
	plotter.Register(NewCpuAgent())
	return plotter
}

func (self *Plotter) Start() {
	atomic.StoreInt32(&self.ploting, 1)
	self.worker.start()
}

func (self *Plotter) Stop() {
	if (atomic.LoadInt32(&self.ploting)) == 0 {
		return
	}
	self.worker.stop()
	atomic.StoreInt32(&self.ploting, 0)
}

func (self *Plotter) Register(agent Agent) {
	if self.Ploting() {
		agent.Start()
	}
	self.worker.register(agent)
}

func (self *Plotter) Unregister(agent Agent) {
	self.worker.unregister(agent)
}

func (self *Plotter) Ploting() bool {
	return atomic.LoadInt32(&self.ploting) > 0
}

func (self *Plotter) SetSeed(addr common.Address) {
	self.seed = addr
}

func (self *Plotter) GetSeed() common.Address {
	return self.seed
}

func (self *Plotter) SetPlotdataDir(dir string) {
	self.plotdataDir = dir
}

func (self *Plotter) GetPlotdataDir() string {
	return self.plotdataDir
}

func (self *Plotter) CommitWork(start uint64, quantity uint64) (tid string, err error) {
	if atomic.LoadInt32(&self.ploting) != 1 {
		return emptyTaskId, errPlotterNotStarted
	}
	os.MkdirAll(self.plotdataDir, 0775)
	work := &Work{
		coinbase:      self.seed,
		plotdataDir:   self.plotdataDir,
		startNonce:    start,
		nonceQuantity: quantity,
	}
	filePath := work.calcPlotfilePath()
	if _, err := os.Stat(filePath); err == nil {
		return work.taskId(), nil
	}

	return self.worker.push(work)
}

func (self *Plotter) GetTaskIds() []string {
	return self.worker.getTaskIds()
}

func (self *Plotter) GetTaskProgress(tid string) (rate int64, err error) {
	paths := []string{self.plotdataDir, tid}
	filePath := strings.Join(paths, "/")
	if _, err := os.Stat(filePath); err == nil {
		return 100, nil
	}

	return self.worker.getTaskProgress(tid)
}

func (self *Plotter) RemovePlotDataById(id string) (err error) {
	s := []string{self.plotdataDir, id}
	filePath := strings.Join(s, "/")
	log.Info("RemovePlotDataById", "filePath", filePath)

	err = os.Remove(filePath)
	return err
}

func (self *Plotter) GetPlotDatalist() (list []string, err error) {
	dataPath := self.plotdataDir
	infos, err := ioutil.ReadDir(dataPath)
	if err != nil {
		return list, err
	}

	for i := 0; i < len(infos); i++ {
		list = append(list, infos[i].Name())
	}

	return list, nil
}

func (self *Plotter) ClearPlotData() error {
	removePath := self.plotdataDir
	log.Info("ClearPlotData", "removePath", removePath)
	err := os.RemoveAll(removePath)
	os.MkdirAll(removePath, 0775)

	return err
}

func (self *Plotter) HavePlotData(addr common.Address) error {
	log.Error("UNSUPPORT")
	return nil
}
