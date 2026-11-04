package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

func clock(id uintptr) float64 {
	var ts syscall.Timespec
	syscall.Syscall(syscall.SYS_CLOCK_GETTIME, id, uintptr(unsafe.Pointer(&ts)), 0)
	return float64(ts.Sec) + float64(ts.Nsec)/1e9
}

func main() {
	start := time.Now()
	r0, m0, b0 := clock(0), clock(1), clock(7) // REALTIME, MONOTONIC, BOOTTIME
	fmt.Println("i  wall(time.Now)  time.Since  dREAL    dMONO    dBOOT    MONO absoluto")
	for i := 1; i <= 10; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("%2d %s  %-10v  %7.3f  %7.3f  %7.3f  %.3f\n",
			i, time.Now().Format("15:04:05.000"),
			time.Since(start).Round(time.Millisecond),
			clock(0)-r0, clock(1)-m0, clock(7)-b0, clock(1))
	}
}