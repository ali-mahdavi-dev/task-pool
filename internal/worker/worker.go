package worker

import (
	"context"
)

type Worker[T any] interface {
	Run(ctx context.Context)
	Shutdown()
	handle(ctx context.Context, command T)
}
