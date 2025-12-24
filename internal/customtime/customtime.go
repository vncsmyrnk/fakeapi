package customtime

import "time"

//go:generate mockgen -source=customtime.go -destination=mock/customtime.go -package=customtime
type Provider interface {
	Sleep(d time.Duration)
}

type RealProvider struct{}

func (RealProvider) Sleep(d time.Duration) {
	time.Sleep(d)
}
