package worklist

// Entry
// Empty path indicates that there are no more jobs to be done
type Entry struct {
	Path string
}

// WorkQueue represents a worklist of jobs to be processed.
type WorkQueue struct {
	jobs chan Entry
}

// Add adds a new job to the worklist.
func (w *WorkQueue) Add(work Entry) {
	w.jobs <- work
}

// Next retrieves the next job from the worklist.
func (w *WorkQueue) Next() Entry {
	j := <-w.jobs
	return j
}

// New creates a new WorkQueue with the specified buffer size.
func New(bufSize int) WorkQueue {
	return WorkQueue{make(chan Entry, bufSize)}
}

// NewJob creates a new job with the given path.
func NewJob(path string) Entry {
	return Entry{path}
}

// Finalize adds a "NoMoreJobs" message to the worklist for each worker.
// Once a worker receives this message, it terminates. After all workers terminate,
// the program can continue.
func (w *WorkQueue) Finalize(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		w.Add(Entry{""})
	}
}
