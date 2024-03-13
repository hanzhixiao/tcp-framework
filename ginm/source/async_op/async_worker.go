package async_op

import "mmo/ginm/zlog"

type AsyncWorker struct {
	taskQue chan func()
}

func (w *AsyncWorker) process(asyncOp func()) {
	if asyncOp == nil {
		zlog.Error("Async operation is empty.")
		return
	}
	if w.taskQue == nil {
		zlog.Error("Task queue has not been initialized.")
		return
	}
	w.taskQue <- func() {
		defer func() {
			if err := recover(); err != nil {
				zlog.Errorf("async process panic: %v", err)
			}
			asyncOp()
		}()
	}
}

func (w *AsyncWorker) loopExecTask() {
	if w.taskQue == nil {
		zlog.Error("Task queue has not been initialized.")
		return
	}
	for {
		task := <-w.taskQue
		if task != nil {
			task()
		}
	}
}
