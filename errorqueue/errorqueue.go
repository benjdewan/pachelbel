package errorqueue

import (
	"fmt"

	"github.com/golang-collections/go-datastructures/queue"
)

// ErrorQueue is a threadsafe queue for storing errors during parallel'
// processing
type ErrorQueue struct {
	q *queue.Queue
}

// New provisions a new ErrorQueue. It should be disposed of using Flush()
func New() *ErrorQueue {
	return &ErrorQueue{q: queue.New(0)}
}

// Enqueue is a variadic function for adding n-many errors to the queue.
func (q *ErrorQueue) Enqueue(items ...error) {
	for _, item := range items {
		if err := q.q.Put(item); err != nil {
			// This only happens if we are using a Queue after Dispose()
			// has been called
			panic(err)
		}
	}
}

// Flush coalesces all the errors in the queue into a single error, disposes of
// the underlying queueing structures and returns the error. If the queue is
// empty, it returns nil
func (q *ErrorQueue) Flush() error {
	if q.q.Empty() {
		q.q.Dispose()
		return nil
	}
	length := q.q.Len()
	errors, qErr := q.q.Get(length)
	if qErr != nil {
		// Get() only returns an error if Dispose has already been called on this
		// queue.
		panic(qErr)
	}
	q.q.Dispose()
	return fmt.Errorf("%d error(s) occurred:\n%v", length, errors)
}
