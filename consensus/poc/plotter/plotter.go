package plotter

import (
	"github.com/pocethereum/pochain/log"
)

/*
 * Plotter是指P盘器，
 *    一个进程中的P盘器掌控着所有的P盘信息，
 *    P盘器是所有P盘实际操作的入口，
 *    P盘器的管理以Work为单位
 */
type Plotter struct {
	workermap map[string]*Worker
	storage   PloterStorage
	allocator PloterAllocator
}

type PloterStorage interface {
	GetAllPlotWorks() []Work
}

type PloterAllocator interface {
	NonceAllocate(id, plotSeed string, plotsize uint64) (startNonce uint64, nonceQuantity uint64, err error)
}

var (
	ploter *Plotter
)

func GetPlotterInstance(storage PloterStorage, allocator PloterAllocator) *Plotter {
	if ploter == nil {
		ploter = newPlotter(storage, allocator)
	} else if storage != nil && storage != ploter.storage {
		panic("storage must be the same")
	}
	return ploter
}

func newPlotter(storage PloterStorage, allocator PloterAllocator) (plotter *Plotter) {
	return &Plotter{
		storage:   storage,
		allocator: allocator,
		workermap: map[string]*Worker{},
	}
}

func (plotter *Plotter) Start() {
	ploter.Reload()
}

func (plotter *Plotter) Stop() {
	var anyonedone = false
	for id, worker := range plotter.workermap {
		log.Info("Stop plot", "Id", id)
		worker.Stop()
		anyonedone = true
	}
	log.Info("Stop plot success", "anyonedone", anyonedone)
}

func (plotter *Plotter) IsPlotting() bool {
	for id, worker := range plotter.workermap {
		if worker.IsWorking() {
			log.Info("Is Plotting true", "Id", id)
			return true
		}
	}
	log.Info("Is Plotting false", "len(workermap)", len(plotter.workermap))
	return false
}

func (plotter *Plotter) Progress() uint {
	var (
		totalSize = uint64(0)
		doneSize  = uint64(0)
	)
	for _, worker := range plotter.workermap {
		progress, plotsize := worker.Progress()
		totalSize += plotsize
		doneSize += uint64(progress) * plotsize / PROGRESS_MAX
	}
	return uint(doneSize * PROGRESS_MAX / totalSize)
}

func (plotter *Plotter) Reload() {
	works := plotter.storage.GetAllPlotWorks()
	for _, w := range plotter.workermap {
		w.Stop()
	}
	for _, w := range works {
		if worker, ok := plotter.workermap[w.Id]; ok {
			//以前有
			worker.Stop()
		}
		//以前没有 或者 有被杀掉
		s, n, e := ploter.allocator.NonceAllocate(w.Id, w.PlotSeed, w.PlotSize)
		if e != nil {
			log.Error("ploter.allocator.NonceAllocate failed", "error", e.Error(), "work", w)
			continue
		}
		newwork := &Work{
			Id:            w.Id,
			PlotSeed:      w.PlotSeed,
			PlotDir:       w.PlotDir,
			PlotSize:      w.PlotSize,
			startNonce:    s,
			nonceQuantity: n,
		}
		newwork.Init()
		newworker := NewWorker(newwork)
		plotter.workermap[w.Id] = newworker
		newworker.Start()
	}
}
