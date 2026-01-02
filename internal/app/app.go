package app

import "context"

type Application interface {
	Initialize(ctx context.Context) error
	Exec(ctx context.Context) error
}
