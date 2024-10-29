package main

import (
	"fmt"
	"time"
)

func orFunc(channels ...<-chan interface{}) <-chan interface{} {
	orChan := make(chan interface{})

	for i := range channels {
		go func() {
			orChan <- (<-channels[i])
		}()
	}

	return orChan

}

func main() {
	var or func(channels ...<-chan interface{}) <-chan interface{} = orFunc

	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}
	start := time.Now()

	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Printf("done after %v\n", time.Since(start))
}
