package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/ratelimiter"
)

func main() {
	var wg sync.WaitGroup
	limiter := ratelimiter.SmoothBuilder[any](5, time.Second).
		WithMaxWaitTime(time.Second).
		OnRateLimitExceeded(func(e failsafe.ExecutionEvent[any]) {
			fmt.Println("rate limit exceeded")
			wg.Done()
		}).
		Build()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			err := failsafe.Run(func() error {
				defer wg.Done()
				fmt.Println("service A")
				return nil
			}, limiter)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	wg.Wait()
}
