package kubesleep

import (
	"fmt"
	"log/slog"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type RepeaterError struct {
	count int
	err   error
}

type repeater struct {
	startingDelay time.Duration
	backoffFactor int
	tries         int
}

func (e RepeaterError) Error() string {
	slog.Error("JOHANNES was here")
	return fmt.Sprintf("operation failed after %d tries %v", e.count, e.err)
}

func (e RepeaterError) Unwrap() error { return e.err }

func repeat(repeatable func() error) error {
	const (
		startingDelay = 100 * time.Millisecond
		backoffFactor = 2
		tries         = 5
	)

	delay := startingDelay
	for i := 1; i <= tries; i++ {
		err := repeatable()
		if err == nil || !apierrors.IsConflict(err) {
			return err
		}
		if i == tries {
			return RepeaterError{i, err}
		}

		slog.Warn("Operation failed and will be retried", "attempt", i, "error", err)
		time.Sleep(delay)
		delay *= backoffFactor
	}
	panic("impossible repeater state")
}
