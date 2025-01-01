package time

import "time"

//go:generate mockgen -source=time.go -destination=mock/time.go -package=time
type Provider interface {
	Sleep(d time.Duration)
}

type RealProvider struct{}

func (RealProvider) Sleep(d time.Duration) {
	time.Sleep(d)
}
