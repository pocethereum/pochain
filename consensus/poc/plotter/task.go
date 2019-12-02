package plotter

import (
	"path/filepath"
	"strconv"
	"strings"
)

type Task struct {
	work          *Work
	startNonce    uint64
	nonceQuantity uint64
	plotfilePath  string
	plotfileSize  uint64
	progress      uint
}

type TASK_STATUS int

const (
	TASK_STATUS_ERROR    = 0
	TASK_STATUS_NONE     = 1
	TASK_STATUS_PLOTTING = 2
	TASK_STATUS_DONE     = 3
)

const (
	PROGRESS_MAX = 10000
)

func NewTask(work *Work, startNonce uint64, nonceQuantity uint64) (task *Task) {
	task = &Task{
		work:          work,
		startNonce:    startNonce,
		nonceQuantity: nonceQuantity,
		progress:      0,
	}
	lowerSeed := strings.ToLower(task.work.PlotSeed)
	if strings.HasPrefix(lowerSeed, "0x") {
		lowerSeed = lowerSeed[2:]
	}

	s := []string{
		lowerSeed,
		strconv.FormatUint(task.startNonce, 10),
		strconv.FormatUint(task.nonceQuantity, 10),
	}
	task.plotfilePath = filepath.Join(task.work.PlotDir, strings.Join(s, "_"))
	task.plotfileSize = task.nonceQuantity << 18

	return task
}

func (t *Task) Check() {

}

func JoinTasks(arr1 []*Task, arr2 []*Task) (retarr []*Task) {
	for _, a := range arr1 {
		retarr = append(retarr, a)
	}
	for _, a := range arr2 {
		retarr = append(retarr, a)
	}
	return retarr
}

func RemoveTask(tasks []*Task, i int) (newtasks []*Task) {
	if i <= 0 || len(tasks) == 0 || i >= len(tasks) {
		return
	}
	if i == 0 {
		newtasks = tasks[1:]
	} else if i == len(tasks)-1 {
		newtasks = tasks[:len(tasks)-1]
	} else {
		newtasks = JoinTasks(tasks[0:i], tasks[i+1:])
	}
	return newtasks
}
