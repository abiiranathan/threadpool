package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/abiiranathan/threadpool"
)

func fetchPage(url string, res chan<- []byte) func(context.Context) error {
	return func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Process the response if needed...
		fmt.Println("Fetched url ", url)
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		res <- data
		return nil
	}
}

func main() {
	// Create a thread pool with a maximum of 3 workers
	pool := threadpool.NewThreadPool(5, threadpool.StopOnError(true))

	// List of URLs to fetch
	urls := []struct {
		URL string
		Res chan []byte
	}{
		{"https://example.com/page1", make(chan []byte)},
		{"https://example.com/page2", make(chan []byte)},
		{"https://example.com/page3", make(chan []byte)},
		{"https://example.com/page4", make(chan []byte)},
		{"https://example.com/page5", make(chan []byte)},
	}

	// Add tasks to the thread pool
	for _, url := range urls {
		pool.AddTask(fetchPage(url.URL, url.Res))
	}

	// Read the results
	for _, url := range urls {
		fmt.Println(string(<-url.Res))
	}

	// Wait for all tasks to complete
	err := pool.Wait()

	// Check for errors
	if err != nil {
		log.Printf("Error encountered: %v", err)
	}
}
