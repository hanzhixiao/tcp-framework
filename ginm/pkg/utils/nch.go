package utils

import (
	"mmo/ginm/source/inter"
	"sync"
)

type nChanel struct {
	Cond *sync.Cond
	//WriteCond *sync.Cond
	taskQueue []inter.Request
	sync.Mutex
	start, end, length, cap int
}

func NewnChanel(cap int) *nChanel {
	return &nChanel{
		Cond: sync.NewCond(&sync.Mutex{}),
		//WriteCond: sync.NewCond(&sync.Mutex{}),
		taskQueue: make([]inter.Request, cap),
		start:     0,
		end:       0,
		cap:       cap,
		length:    0,
		Mutex:     sync.Mutex{},
	}
}

func (ch *nChanel) Store(requests []inter.Request) {
	requestLen := len(requests)
	ch.Cond.L.Lock()
	storedNum := min(requestLen, ch.cap-ch.length)
	ch.length += storedNum
	fkEndPos := ch.end + storedNum
	endpos := (ch.end + storedNum) % ch.cap
	if fkEndPos != endpos {
		tailSize := ch.cap - ch.end
		copy(ch.taskQueue[ch.end:], requests[:tailSize])
		copy(ch.taskQueue[:storedNum-tailSize], requests[tailSize:storedNum])
	} else {
		copy(ch.taskQueue[ch.end:ch.end+storedNum], requests[:storedNum])
	}
	ch.end = endpos
	if storedNum > 0 {
		ch.Cond.Signal()
	}
	if storedNum < requestLen {
		ch.Cond.Wait()
		ch.Cond.L.Unlock()
		ch.Store(requests[storedNum:])
		return
	}
	ch.Cond.L.Unlock()
}

func (ch *nChanel) GetLength() int {
	ch.Cond.L.Lock()
	defer ch.Cond.L.Unlock()
	return ch.length
}

func (ch *nChanel) Load(n int) []inter.Request {
	ch.Cond.L.Lock()
	chLen := ch.length
	loadNum := min(chLen, n)
	requests := make([]inter.Request, 0, n)
	startN := ch.start + loadNum
	startPos := (startN) % ch.cap
	ch.length -= loadNum
	unreadNum := n - loadNum
	if startN != startPos {
		requests = append(requests, ch.taskQueue[ch.start:]...)
		requests = append(requests, ch.taskQueue[:startPos]...)
	} else {
		requests = append(requests, ch.taskQueue[ch.start:ch.start+loadNum]...)
	}
	ch.start = startPos
	if loadNum > 0 {
		ch.Cond.Signal()
	}
	if loadNum < n {
		ch.Cond.Wait()
		ch.Cond.L.Unlock()
		requests = append(requests, ch.Load(unreadNum)...)
		return requests
	}
	ch.Cond.L.Unlock()
	return requests
}

func (ch *nChanel) LoadHalf() []inter.Request {
	return ch.Load(ch.GetLength() / 2)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
