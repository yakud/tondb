package main

import (
	"context"
	"fmt"
	"sync"
)

func main() {
	stopChan := make(chan struct{})
	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-stopChan:
		case <-ctx.Done():
			fmt.Println("goroutine stopped")
		}
	}()

	go func() {
		stopChan <- struct{}{}
	}()

	wg.Wait()
	fmt.Println("done!")
}
