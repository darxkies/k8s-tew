package utils

type Limiter struct {
	channel chan bool
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{channel: make(chan bool, limit)}
}

func (limiter Limiter) Lock() {
	limiter.channel <- true
}

func (limiter Limiter) Unlock() {
	<-limiter.channel
}
