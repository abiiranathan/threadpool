package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/abiiranathan/threadpool"
)

func main() {
	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			// Query 1
			time.Sleep(time.Second)
			fmt.Println("Query 1 executed")
			return nil
		},
		func(ctx context.Context) error {
			// Query 2
			time.Sleep(time.Second)
			fmt.Println("Query 2 executed")
			return nil
		},
		func(ctx context.Context) error {
			// Query 3
			time.Sleep(time.Second)
			fmt.Println("Query 3 executed")
			return nil
		},
		func(ctx context.Context) error {
			// Query 4
			time.Sleep(time.Second)
			fmt.Println("Query 4 executed")
			return nil
		},
		func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				fmt.Println("context timeout, skipping query 5")
				return nil
			case <-time.After(time.Second * 2):
				fmt.Println("Query 5 still running")
				return fmt.Errorf("timeout")
			}
		},
	}

	// Create a thread pool.
	pool := threadpool.NewThreadPool(5,
		threadpool.StopOnError(false),
		threadpool.Timeout(time.Second),
	)

	pool.Start()

	start := time.Now()

	// Add tasks to the thread pool
	for _, query := range tasks {
		pool.AddTask(query)
	}

	// Wait for all tasks to complete
	err := pool.Wait()
	if err != nil {
		log.Println(err)
	}

	took := time.Since(start)
	fmt.Println("All queries completed in ", took.Seconds(), " seconds")
}
