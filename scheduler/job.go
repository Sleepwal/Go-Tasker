package scheduler

import "context"

type Job interface {
	ID() string
	Run(ctx context.Context) error
}
