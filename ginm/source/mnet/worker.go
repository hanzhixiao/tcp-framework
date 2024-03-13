package mnet

import "mmo/ginm/source/inter"

type worker struct {
	workerId int
	que      chan inter.Request
}

func (w *worker) GetWorkerId() int {
	return w.workerId
}

func (w *worker) GetRequestQueue() chan inter.Request {
	return w.que
}

func NewWorker(chanSize int, workerId int) inter.Worker {
	return &worker{que: make(chan inter.Request, chanSize), workerId: workerId}
}
