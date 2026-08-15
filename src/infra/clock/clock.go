package clock

import "time"

// SystemClock satisfies usecase.Clock via the wall clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}
