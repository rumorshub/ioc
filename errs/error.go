package errs

import (
	"context"
	"errors"
	"fmt"
)

var Success = errors.New("SUCCESS")

func IsSuccess(err error) bool {
	return err == nil || errors.Is(err, Success)
}

func Go(err error) error {
	if IsSuccess(err) {
		return nil
	}
	return err
}

func Append(left error, right error) error {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	}

	return fmt.Errorf("%w; %w", left, right)
}

func IsCanceledOrDeadline(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
