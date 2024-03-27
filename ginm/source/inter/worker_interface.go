package inter

import (
	"time"
)

type Worker interface {
	GetRequestQueue() chan Request
	GetWorkerId() int
	GetRobQueue() Nch
	IsTimeToRob() bool
	SetLastFailTime(failTime time.Time)
}
