package time

import "time"

//go:generate mockgen -source=time.go -destination=mock/time.go -package=time
type TimeProvider interface {
	Sleep(d time.Duration)
}

type RealTimeProvider struct{}

func (RealTimeProvider) Sleep(d time.Duration) {
	time.Sleep(d)
}
