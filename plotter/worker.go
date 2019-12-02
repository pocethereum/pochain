package plotter

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pocethereum/pochain/common"
)

const (
	maxTaskQueueSize   = 10000
	maxResultQueueSize = 10
	emptyTaskId        = ""
)

type Work struct {
	coinbase      common.Address
	plotdataDir   string
	startNonce    uint64
	nonceQuantity uint64
	plotfilePath  string
	progress      int
}

func (self *Work) calcPlotfilePath() string {
	s := []string{
		strings.ToLower(self.coinbase.Hex())[2:],
		strconv.FormatUint(self.startNonce, 10),
		strconv.FormatUint(self.nonceQuantity, 10),
	}
	self.plotfilePath = filepath.Join(self.plotdataDir, strings.Join(s, "_"))
	return self.plotfilePath
}

func (self *Work) taskId() string {
	return filepath.Base(self.plotfilePath)
}

type Result struct {
	work *Work
	err  error
}

type ProgressResult struct {
	rate int64
	err  error
}

type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *Result)
	Start()
	Stop()
}

type worker struct {
	mu sync.Mutex
	wg sync.WaitGroup

	workCh chan *Work
	recvCh chan *Result
	stopCh chan struct{}
	agents map[Agent]struct{}

	taskMap       map[string]*Work
	taskProgress  map[string]*ProgressResult
	taskQueueSize int32
	atWork        int32
	ploting       int32
}

func newWorker() *worker {
	worker := &worker{
		workCh:       make(chan *Work, maxTaskQueueSize),
		recvCh:       make(chan *Result, maxResultQueueSize),
		stopCh:       make(chan struct{}, 1),
		agents:       make(map[Agent]struct{}),
		taskMap:      make(map[string]*Work),
		taskProgress: make(map[string]*ProgressResult),
	}
	return worker
}

func (self *worker) start() {
	self.mu.Lock()
	defer self.mu.Unlock()

	atomic.StoreInt32(&self.ploting, 1)
	for agent := range self.agents {
		agent.Start()
	}
	go self.update()
}

func (self *worker) randomAgent() Agent {
	self.mu.Lock()
	defer self.mu.Unlock()

	for agent := range self.agents {
		return agent
	}

	agent := NewCpuAgent()
	agent.Start()
	agent.SetReturnCh(self.recvCh)
	self.agents[agent] = struct{}{}
	return agent
}

func (self *worker) update() {
	ticker := time.NewTicker(time.Millisecond * 5000)
	for {
		select {
		case <-self.stopCh:
			return
		case <-ticker.C:
			if atomic.LoadInt32(&self.atWork) == 0 {
				agent := self.randomAgent()
				func() {
					self.mu.Lock()
					defer self.mu.Unlock()
					if atomic.LoadInt32(&self.taskQueueSize) > 0 {
						work := <-self.workCh
						agent.Work() <- work
						atomic.AddInt32(&self.taskQueueSize, -1)
						atomic.AddInt32(&self.atWork, 1)
					}
				}()
			}
		case result := <-self.recvCh:
			func() {
				self.mu.Lock()
				defer self.mu.Unlock()
				if atomic.LoadInt32(&self.atWork) > 0 {
					atomic.AddInt32(&self.atWork, -1)
				}
				tid := result.work.taskId()
				delete(self.taskMap, tid)
				if _, ok := self.taskProgress[tid]; ok {
					self.taskProgress[tid].rate = 100
					self.taskProgress[tid].err = result.err
				}
			}()
		}
	}
}

func (self *worker) stop() {
	self.wg.Wait()

	self.mu.Lock()
	defer self.mu.Unlock()

	if atomic.LoadInt32(&self.ploting) == 1 {
		for agent := range self.agents {
			agent.Stop()
		}
	}
	atomic.StoreInt32(&self.ploting, 0)
	atomic.StoreInt32(&self.atWork, 0)
	atomic.StoreInt32(&self.taskQueueSize, 0)

	self.taskMap = make(map[string]*Work)
	self.taskProgress = make(map[string]*ProgressResult)

	self.stopCh <- struct{}{}
done:
	for {
		select {
		case <-self.workCh:
		case <-self.recvCh:
		default:
			break done
		}
	}
}

func (self *worker) register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.recvCh)
}

func (self *worker) unregister(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.agents, agent)
	agent.Stop()
}

func (self *worker) push(work *Work) (tid string, err error) {
	work.calcPlotfilePath()
	taskId := work.taskId()

	self.mu.Lock()
	defer self.mu.Unlock()

	if _, ok := self.taskMap[taskId]; ok {
		return taskId, errTaskAlreadyInQueue
	}

	if self.taskQueueSize >= maxTaskQueueSize-1 {
		return taskId, errTaskQueueOverflow
	}

	self.taskMap[taskId] = work
	self.taskProgress[taskId] = &ProgressResult{}
	atomic.AddInt32(&self.taskQueueSize, 1)
	self.workCh <- work
	return taskId, nil
}

func (self *worker) getTaskIds() []string {
	self.mu.Lock()
	defer self.mu.Unlock()

	tids := []string{}
	for tid := range self.taskProgress {
		tids = append(tids, tid)
	}
	return tids
}

func (self *worker) getTaskProgress(tid string) (rate int64, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if progRet, ok := self.taskProgress[tid]; ok {
		work, _ := self.taskMap[tid]
		progRet.rate = int64(work.progress)
		return progRet.rate, progRet.err
	}
	return 0, errTaskNotFound
}
