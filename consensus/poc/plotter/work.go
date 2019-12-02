package plotter

import ()

/*
 *一个Work对应着某一个需要P的目录;
 *一个Work由多个Task组成;
 *    每一个Work有一个work进度，由已完成任务及正在进行的任务共同构成;
 *    每一个Task代表一个即将进行或者正在进行的P盘操作。
 **/
type Work struct {
	Id            string
	PlotSeed      string
	PlotDir       string
	PlotSize      uint64
	startNonce    uint64
	nonceQuantity uint64
	todoTasks     []*Task
	doingTasks    []*Task
	doneTasks     []*Task
}

func (work *Work) Init() {
	const ONE_TASK_NONCE_NUM = 4096
	work.todoTasks = []*Task{}
	work.doingTasks = []*Task{}
	work.doneTasks = []*Task{}
	for i := uint64(0); i < work.nonceQuantity; i += ONE_TASK_NONCE_NUM {
		task := NewTask(work, work.startNonce+i, ONE_TASK_NONCE_NUM)
		work.todoTasks = append(work.todoTasks, task)
	}
}

func (work *Work) GetTask() (task *Task) {
	if len(work.todoTasks) == 0 {
		return nil
	}
	task = work.todoTasks[0]
	work.todoTasks = work.todoTasks[1:]
	work.doingTasks = append(work.doingTasks, task)
	return task
}

func (work *Work) CommitTask(task *Task) {
	for i, doing := range work.doingTasks {
		if doing.startNonce == task.startNonce {
			work.doneTasks = append(work.doneTasks, doing)
			work.doingTasks = RemoveTask(work.doingTasks, i)
		}
	}
	return
}

func (work *Work) RollbackTask(task *Task) {

	return
}

func (work *Work) Progress() uint {
	total := (len(work.todoTasks) + len(work.doingTasks) + len(work.doneTasks))
	done := len(work.doneTasks) * PROGRESS_MAX
	for _, doing := range work.doingTasks {
		done += int(doing.progress)
	}
	return uint(done / total)
}
