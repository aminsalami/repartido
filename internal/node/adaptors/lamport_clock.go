package adaptors

import "sync"

// LamportClock is a partial-order clock implementing ports.ILamportClock
type LamportClock struct {
	counter uint64
	sync.Mutex
}

func NewLamportClock() *LamportClock {
	return &LamportClock{counter: 0}
}

func (l *LamportClock) Witness(t uint64) bool {
	l.Lock()
	defer l.Unlock()
	if t > l.counter {
		//l.counter = t + 1	 --> Is it really necessary to increment it?
		l.counter = t
		return true
	}
	return false
}

func (l *LamportClock) Time() uint64 {
	return l.counter
}

func (l *LamportClock) Tick() uint64 {
	l.Lock()
	defer l.Unlock()
	l.counter = l.counter + 1
	return l.counter
}
