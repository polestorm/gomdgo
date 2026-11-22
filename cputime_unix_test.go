//go:build unix

package gomdgo

import (
	"syscall"
	"time"
)

// cpuTime zwraca czas procesora zużyty dotąd przez proces (użytkownika i systemu). W odróżnieniu od zegara
// nie rośnie, gdy proces czeka na rdzeń, więc pomiar nie zależy od obciążenia maszyny.
func cpuTime() time.Duration {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return time.Duration(time.Now().UnixNano())
	}
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}
