package plotter

import "errors"

var (
	errPlotterNotStarted      = errors.New("plotter not started")
	errTaskQueueOverflow      = errors.New("task queue overflow")
	errPlotDataAlreadyExisted = errors.New("plotdata already existed")
	errTaskAlreadyInQueue     = errors.New("task already in queue")
	errTaskNotFound           = errors.New("task not found")
)
