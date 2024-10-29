package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

func PrintTime() {
	curTime, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occured: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Current time according to the ntp:        ", curTime)
	fmt.Println("Current time according to the local clock:", time.Now())
}

func main() {
	PrintTime()
}
