package plotter

import (
	plotpoc "github.com/pocethereum/pochain/consensus/poc/plot"
	"github.com/pocethereum/pochain/log"
	plotparams "github.com/pocethereum/pochain/params/plot"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type Worker struct {
	lock      sync.Mutex
	isworking uint64
	work      *Work
}

var (
	errOk        error = nil
	errFileError       = errors.New("File operate error")
)

func NewWorker(work *Work) (worker *Worker) {
	worker = &Worker{
		work:      work,
		lock:      sync.Mutex{},
		isworking: 0,
	}
	return worker
}

func (w *Worker) Start() {
	go w.working()
}

func (w *Worker) Stop() {
	if atomic.LoadUint64(&w.isworking) == 0 {
		return
	}
	atomic.StoreUint64(&w.isworking, 0)
}

func (w *Worker) IsWorking() bool {
	return atomic.LoadUint64(&w.isworking) == 1
}

func (w *Worker) Progress() (progress uint, plotsize uint64) {
	return w.work.Progress(), w.work.PlotSize
}

func (w *Worker) working() {
	if atomic.LoadUint64(&w.isworking) == 1 {
		return
	}
	atomic.StoreUint64(&w.isworking, 1)

	for atomic.LoadUint64(&w.isworking) == 1 {
		// Step 1. Get Task
		task := w.work.GetTask()
		if task == nil {
			log.Info("all works done, plotter working exit")
			atomic.StoreUint64(&w.isworking, 0)
			break
		}

		// Step 2. Check Task
		switch status := w.taskCheck(task); status {
		case TASK_STATUS_NONE:
			log.Info("task is new, ready to doPlot", "task", task)
		case TASK_STATUS_DONE:
			log.Info("task have been done, commit & continue next", "task", task)
			w.work.CommitTask(task)
			continue
		case TASK_STATUS_ERROR, TASK_STATUS_PLOTTING:
			log.Info("something error", "status", status)
			w.work.RollbackTask(task)
			continue
		}

		// Step 3. Do Plot
		switch err := w.doPlot(task); err {
		case errOk:
			log.Info("doPlot done, commit & continue next", "task", task)
			w.work.CommitTask(task)
			continue
		default:
			log.Info("something error", "error", err.Error())
			w.work.RollbackTask(task)
			continue
		}
	}
}

func (w *Worker) taskCheck(task *Task) (status TASK_STATUS) {
	if stat, err := os.Stat(task.plotfilePath); os.IsNotExist(err) {
		log.Info("file is not exist", "plotfilePath", task.plotfilePath)
		return TASK_STATUS_NONE
	} else if err != nil {
		log.Info("stat file error,remove it", "plotfilePath", task.plotfilePath)
		os.Remove(task.plotfilePath)
		return TASK_STATUS_NONE
	} else if stat.Size() != int64(task.plotfileSize) {
		log.Info("filesize changed, remove it", "plotfilePath", task.plotfilePath,"stat size", stat.Size(), "task size", task.plotfileSize)
		os.Remove(task.plotfilePath)
		return TASK_STATUS_NONE
	} else {
		return TASK_STATUS_DONE
	}
}

func (w *Worker) doPlot(task *Task) (err error) {
	var (
		nonceIndex     = uint64(0)
		scoopIndex     = uint64(0)
		scoopLimit     = plotparams.ScoopsPerPlot * task.nonceQuantity
		scoopDataBytes = make([]byte, plotparams.ScoopSize)
		origPath       = task.plotfilePath + ".orig"
		destPath       = task.plotfilePath + ".dest"
		origFd, err1   = os.OpenFile(origPath, os.O_RDWR|os.O_CREATE, 0600)
		destFd, err2   = os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0600)
	)
	if err1 != nil {
		log.Error("OpenFile error", "path", origPath)
		return errFileError
	} else if err2 != nil {
		log.Error("OpenFile error", "path", destPath)
		return errFileError
	}
	defer origFd.Close()
	defer destFd.Close()
	defer os.Remove(origPath)
	defer os.Remove(destPath)

	// Step 1. Plot of GENERATE
	step := "GENERATE"
	progress := uint(0)
	progress_rate := uint64(PROGRESS_MAX / 2)
	log.Info("Plot doing", "step", step)
	for atomic.LoadUint64(&w.isworking) == 1 && nonceIndex < task.nonceQuantity {
		nonce := task.startNonce + nonceIndex
		lowerSeed := strings.ToLower(task.work.PlotSeed)
		if strings.HasPrefix(lowerSeed, "0x") {
			lowerSeed = lowerSeed[2:]
		}
		mp := plotpoc.NewMiningPlot(lowerSeed, nonce)
		if _, err := origFd.Write(mp.Data()); err != nil {
			log.Info("origFd.Write error", "error", err.Error())
			return errFileError
		}
		nonceIndex++
		task.progress = progress + uint(nonceIndex*progress_rate/task.nonceQuantity)
	}

	// Step 2. Plot of OPTIMIZE
	step = "OPTIMIZE"
	progress += uint(progress_rate)
	progress_rate = uint64(PROGRESS_MAX / 2)
	log.Info("Plot doing", "step", step)
	for atomic.LoadUint64(&w.isworking) == 1 && scoopIndex < scoopLimit {
		s := scoopIndex / task.nonceQuantity
		n := scoopIndex % task.nonceQuantity
		index := n*plotparams.PlotSize + s*plotparams.ScoopSize
		if _, err := origFd.Seek(int64(index), 0); err != nil {
			log.Info("origFd.Seek error", "error", err.Error())
			return errFileError
		}
		if _, err := origFd.Read(scoopDataBytes); err != nil {
			log.Info("origFd.Read error", "error", err.Error())
			return errFileError
		}
		if _, err := destFd.Write(scoopDataBytes); err != nil {
			log.Info("origFd.Write error", "error", err.Error())
			return errFileError
		}

		task.progress = progress + uint(nonceIndex*progress_rate/task.nonceQuantity)
		scoopIndex++
	}

	// Step 3. Rename to destination file
	log.Info("ploting done, rename file", "destPath", destPath, "work", w.work)
	destFd.Close()
	err = os.Rename(destPath, task.plotfilePath)
	if err != nil {
		log.Info("Rename error", "error", err, "destPath", destPath, "work.plotfilePath", task.plotfilePath)
		return errFileError
	} else {
		log.Info("Rename success & plot success", "destPath", destPath, "work.plotfilePath", task.plotfilePath)
	}

	return err
}
