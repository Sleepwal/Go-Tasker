package scheduler

type Retry interface {
	Retry(fn func() error) error
}
