package testautomation

import (
	"time"
)

type TestClock struct {
	Offset time.Duration
}

func (c *TestClock) Now() time.Time {
	return time.Now().UTC().Add(c.Offset)
}

func (c *TestClock) SetTimeOffset(d time.Duration) {
	c.Offset = d
}

func (c *TestClock) TimeOffset() time.Duration {
	return c.Offset
}
