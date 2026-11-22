//go:build !unix

package gomdgo

import "time"

// cpuTime poza systemami uniksowymi zwraca czas zegara: pomiar działa, ale zależy od obciążenia maszyny.
func cpuTime() time.Duration {
	return time.Duration(time.Now().UnixNano())
}
