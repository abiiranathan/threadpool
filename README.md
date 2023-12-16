# ThreadPool

ThreadPool is a simple Go package that provides a flexible and efficient thread pool implementation, allowing you to parallelize and manage the execution of tasks concurrently. It supports context cancellation, error handling, and timeout functionality.

## Installation

To install the package, use the following go get command:

```bash
go get -u github.com/abiiranathan/threadpool

```

## Usage

Here is an example demonstrating how to use the ThreadPool package to execute a set of tasks concurrently:

```go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/abiiranathan/threadpool"
)

func main() {
	// Define a set of tasks
	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			time.Sleep(time.Second)
			fmt.Println("Query 1 executed")
			return nil
		},
		func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				fmt.Println("Context timeout, skipping Query 5")
				return nil
			case <-time.After(time.Second * 2):
				fmt.Println("Query 5 still running")
				return fmt.Errorf("network timeout")
			}
		},
	}

	// Create a thread pool with a maximum of 5 workers
	pool := threadpool.NewThreadPool(5, true)
	pool.SetTimeout(time.Second)
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
	fmt.Println("All queries completed in", took.Seconds(), "seconds")
}
```

## Features

- **Concurrent Execution**: Execute tasks concurrently with a configurable number of workers.
- **Context Cancellation**: Utilize context cancellation for graceful termination of tasks.
- **Error Handling: Stop** processing new tasks upon encountering the first error, if configured.
- **Timeout Functionality**: Set a deadline for all tasks with the SetTimeout function.
- **WithContext** to accept custom context.

Run Example:

```bash
go run cmd/main.go
```
