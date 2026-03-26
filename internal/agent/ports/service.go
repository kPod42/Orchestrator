package ports

import "context"

type Service interface {
	Name() string
	Start(ctx context.Context) error
	Ready() <-chan struct{}
}
