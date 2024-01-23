package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/Xexanoath/ggrep/worker"
	"github.com/Xexanoath/ggrep/worklist"
	"github.com/alexflint/go-arg"
)

// discoverDirs recursively explores directories and adds files to the worklist.
func discoverDirs(wl *worklist.WorkQueue, path string) {
	// Read the directory entries at the specified path.
	entries, err := os.ReadDir(path)
	if err != nil {
		// Print an error message if ReadDir fails and return from the function.
		fmt.Println("ReadDir error:", err)
		return
	}

	// Iterate over each directory entry.
	for _, entry := range entries {
		// Check if the entry is a directory.
		if entry.IsDir() {
			// If it's a directory, recursively call discoverDirs on the next path.
			nextPath := filepath.Join(path, entry.Name())
			discoverDirs(wl, nextPath)
		} else {
			// If it's a file, add a new job to the worklist with the file's path.
			wl.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}
}

// SearchTerm is the term to search for.
// SearchDir is the directory to search in. If not provided, the current directory is used.
var args struct {
	SearchTerm string `arg:"positional,required,help:the term to search for"`
	SearchDir  string `arg:"positional,help:the directory to search in"`
}

func main() {
	// Parse command line arguments
	arg.MustParse(&args)

	// Wait group for managing worker goroutines
	var workersWg sync.WaitGroup

	// Create a worklist with a buffer of 100 entries
	wl := worklist.New(100)

	// Channel for collecting worker results with a buffer of 100 entries
	results := make(chan worker.Result, 100)

	// Number of worker goroutines
	numWorkers := runtime.NumCPU()

	// Start a goroutine to discover directories and populate the worklist
	workersWg.Add(1)
	go func() {
		defer workersWg.Done()
		discoverDirs(&wl, args.SearchDir)
		wl.Finalize(numWorkers)
	}()

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for {
				// Get the next work entry from the worklist
				workEntry := wl.Next()

				// Check if the path is not empty
				if workEntry.Path != "" {
					// Perform the search operation on the file
					workerResult := worker.FindInFile(workEntry.Path, args.SearchTerm)
					if workerResult != nil {
						// Send each inner result to the results channel
						for _, r := range workerResult.Inner {
							results <- r
						}
					}
				} else {
					// When the path is empty, no more jobs available, so exit the goroutine
					return
				}
			}
		}()
	}

	// Channel to signal when all worker goroutines are done
	blockWorkersWg := make(chan struct{})
	go func() {
		// Wait for all worker goroutines to finish
		workersWg.Wait()
		close(blockWorkersWg)
	}()

	// Wait group for managing display goroutine
	var displayWg sync.WaitGroup

	// Start a goroutine to display results
	displayWg.Add(1)
	go func() {
		for {
			select {
			case r := <-results:
				// Display the result
				fmt.Printf("%v[%v]:%v\n", r.Path, r.LineNum, r.Line)
			case <-blockWorkersWg:
				// Make sure the channel is empty before aborting display goroutine
				if len(results) == 0 {
					displayWg.Done()
					return
				}
			}
		}
	}()
	// Wait for the display goroutine to finish
	displayWg.Wait()
}
