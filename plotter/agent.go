package plotter

import (
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	plotpoc "github.com/pocethereum/pochain/consensus/poc/plot"
	"github.com/pocethereum/pochain/log"
	plotparams "github.com/pocethereum/pochain/params/plot"
)

type CpuAgent struct {
	mu sync.Mutex

	workCh        chan *Work
	stop          chan struct{}
	quitCurrentOp chan struct{}
	returnCh      chan<- *Result

	isPloting int32 // isPloting indicates whether the agent is currently ploting
}

func NewCpuAgent() *CpuAgent {
	return &CpuAgent{
		workCh: make(chan *Work, 1),
		stop:   make(chan struct{}, 1),
	}
}

func (self *CpuAgent) Work() chan<- *Work {
	return self.workCh
}

func (self *CpuAgent) SetReturnCh(ch chan<- *Result) {
	self.returnCh = ch
}

func (self *CpuAgent) Stop() {
	if !atomic.CompareAndSwapInt32(&self.isPloting, 1, 0) {
		return
	}
	self.stop <- struct{}{}
done:
	for {
		select {
		case <-self.workCh:
		default:
			break done
		}
	}
}

func (self *CpuAgent) Start() {
	if !atomic.CompareAndSwapInt32(&self.isPloting, 0, 1) {
		return
	}
	go self.update()
}

// Cancel the current task
func (self *CpuAgent) Cancel() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.quitCurrentOp != nil {
		close(self.quitCurrentOp)
	}
	self.quitCurrentOp = nil
}

func (self *CpuAgent) update() {
out:
	for {
		select {
		case work := <-self.workCh:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
			}
			self.quitCurrentOp = make(chan struct{})
			go self.plot(work, self.quitCurrentOp)
			self.mu.Unlock()
		case <-self.stop:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
				self.quitCurrentOp = nil
			}
			self.mu.Unlock()
			break out
		}
	}
}

func (self *CpuAgent) plot(work *Work, quit <-chan struct{}) {
	result := &Result{
		work: work,
		err:  nil,
	}
	if checkFileExistOrNot(work.plotfilePath) {
		result.err = errPlotDataAlreadyExisted
		self.returnCh <- result
		return
	}

	result.work.progress = 0
	origPath := work.plotfilePath + ".orig"
	destPath := work.plotfilePath + ".dest"
	os.Remove(origPath)
	os.Remove(destPath)

	origFd, err1 := os.OpenFile(origPath, os.O_RDWR|os.O_CREATE, 0600)
	if err1 != nil {
		result.err = err1
		self.returnCh <- result
		return
	}
	defer func() {
		origFd.Close()
		os.Remove(origPath)
	}()

	destFd, err2 := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0600)
	if err2 != nil {
		result.err = err2
		self.returnCh <- result
		return
	}
	defer func() {
		destFd.Close()
		os.Remove(destPath)
	}()

	var (
		optimize       = 0
		nonceIndex     = uint64(0)
		scoopIndex     = uint64(0)
		scoopLimit     = plotparams.ScoopsPerPlot * work.nonceQuantity
		scoopDataBytes = make([]byte, plotparams.ScoopSize)
	)

	log.Info("ploting started", "work", result.work)
plot:
	for {
		select {
		case <-quit:
			log.Info("ploting aborted", "attemps", nonceIndex)
			result.err = errors.New("ploting aborted")
			self.returnCh <- result
			return
		default:
			if optimize == 0 {
				if nonceIndex == work.nonceQuantity {
					optimize = 1
				} else {
					nonce := work.startNonce + nonceIndex
					seed := strings.ToLower(work.coinbase.Hex()[2:])
					mp := plotpoc.NewMiningPlot(seed, nonce)
					_, result.err = origFd.Write(mp.Data())
					if result.err != nil {
						self.returnCh <- result
						return
					}
					nonceIndex++
					result.work.progress = 0 + int(nonceIndex*50/work.nonceQuantity)
				}
			} else {
				if scoopIndex == scoopLimit {
					break plot
				}

				s := scoopIndex / work.nonceQuantity
				n := scoopIndex % work.nonceQuantity
				index := n*plotparams.PlotSize + s*plotparams.ScoopSize
				_, result.err = origFd.Seek(int64(index), 0)
				if result.err != nil {
					self.returnCh <- result
					return
				}

				_, result.err = origFd.Read(scoopDataBytes)
				if result.err != nil {
					self.returnCh <- result
					return
				}

				_, result.err = destFd.Write(scoopDataBytes)
				if result.err != nil {
					self.returnCh <- result
					return
				}
				result.work.progress = 50 + int(scoopIndex*50/scoopLimit)
				scoopIndex++
			}
		}
	}

	log.Info("ploting done, rename file", "destPath", destPath, "work", result.work)
	destFd.Close()
	err := os.Rename(destPath, work.plotfilePath)
	if err != nil {
		log.Info("Rename error", "error", err, "destPath", destPath, "work.plotfilePath", work.plotfilePath)
		result.err = errors.New("Rename error" + err.Error())
	} else {
		log.Info("Rename success & plot success", "destPath", destPath, "work.plotfilePath", work.plotfilePath)
	}
	self.returnCh <- result
}

func checkFileExistOrNot(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
