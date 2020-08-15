package ahoy

import "time"

var _ Clock = new(MockClock)

type MockClock struct {
	NowFn func() time.Time
}

func (m *MockClock) Now() time.Time {
	return m.NowFn()
}
