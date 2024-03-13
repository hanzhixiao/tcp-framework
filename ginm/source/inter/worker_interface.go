package inter

type Worker interface {
	GetRequestQueue() chan Request
	GetWorkerId() int
}
