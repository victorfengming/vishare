package defaults

import "time"

const (
	SendBufSize    = 256
	PingInterval   = 5 * time.Second
	EdgePollMs     = 10 * time.Millisecond
	EdgeHysteresis = 3
	BackoffMin     = 1 * time.Second
	BackoffMax     = 30 * time.Second
)
