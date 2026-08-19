package runner

import (
	"context"
	"hl/ui"
	"io"
)

type Result struct {
	ExitCode int
	Error    error
}
type Runner interface {
	Validate(ctx context.Context, step ui.Step) error

	Snapshot(ctx context.Context, step ui.Step) error

	Run(ctx context.Context, step ui.Step, stdout, stderr io.Writer) Result
}
