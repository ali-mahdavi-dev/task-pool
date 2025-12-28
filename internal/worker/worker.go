package worker

import (
	"context"
)

type Worker[T any] interface {
	Run(ctx context.Context)
	Shutdown(ctx context.Context)
	handle(ctx context.Context, command T)
}
