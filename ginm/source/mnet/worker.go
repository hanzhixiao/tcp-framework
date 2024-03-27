package mnet

import (
	"mmo/ginm/source/inter"
	"time"
)

type worker struct {
	workerId     int
	que          chan inter.Request
	robQue       inter.Nch
	lastFailTime time.Time
}

func (w *worker) GetRobQueue() inter.Nch {
	return w.robQue
}

func (w *worker) GetWorkerId() int {
	return w.workerId
}

func (w *worker) GetRequestQueue() chan inter.Request {
	return w.que
}

func NewWorker(chanSize int, workerId int) inter.Worker {
	return &worker{que: make(chan inter.Request, chanSize), workerId: workerId, robQue: NewnChanel(chanSize)}
}

func (w *worker) IsTimeToRob() bool {
	now := time.Now()
	if now.Sub(w.lastFailTime) > rob_interval {
		return true
	}
	return false
}
func (w *worker) SetLastFailTime(failTime time.Time) {
	w.lastFailTime = failTime
}
