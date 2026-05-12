package client

import "time"

type RetryAfterError struct {
	RetryAfter time.Duration
}

func (e *RetryAfterError) Error() string {
	return "retry after"
}
