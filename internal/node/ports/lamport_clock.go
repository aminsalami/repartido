package ports

type Ltime uint64

type ILamportClock interface {
	// Witness is called to update the local time with the received time from other processes.
	// returns true if witnessed a new time.
	Witness(Ltime) bool
	// Time returns the current local lamport time
	Time() Ltime
	// Tick adds 1 unit to the clock
	Tick() Ltime
}
