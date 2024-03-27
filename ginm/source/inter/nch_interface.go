package inter

type Nch interface {
	Store(requests []Request)
	Length() int
	Cap() int
	Load(n int) []Request
	LoadHalf() []Request
}
