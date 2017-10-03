package runner

import (
	"fmt"
	"sync"
	"time"

	"github.com/benjdewan/pachelbel/connection"
	progressbars "github.com/benjdewan/pachelbel/progress"
	"github.com/golang-collections/go-datastructures/queue"
)

// Accessor is the interface for any Compose Deployment information request.
// To make any deployment mutations (creating new deployments or updating
// existing ones), the provided object must also implement the Deployment
// interface.
type Accessor interface {
	GetName() string
	GetType() string
}

// Controller is the stateful object that handles actually running Runner to
// work with Compose
type Controller struct {
	cxn      *connection.Connection
	progress *progressbars.ProgressBars
	dryRun   bool
}

func NewController(cxn *connection.Connection, dryRun bool) *Controller {
	ctl := &Controller{
		cxn:      cxn,
		progress: progressbars.New(),
		dryRun:   dryRun,
	}
	ctl.progress.RefreshRate = 3 * time.Second
	return ctl
}

// Run processes a slice of Runners. Doing what ever action has been set as
// ther 'run' function in parallel.
func (ctl *Controller) Run(runners []Runner) error {
	runners = ctl.register(runners)

	var wg sync.WaitGroup
	wg.Add(len(runners))
	q := queue.New(0)
	ctl.progress.Start()

	for _, runner := range runners {
		go func(r Runner) {
			if err := r.Run(ctl.cxn, r.Target); err != nil {
				ctl.progress.Error(r.Target.GetName())
				enqueue(q, err)
			} else {
				ctl.progress.Done(r.Target.GetName())
			}
			wg.Done()
		}(runner)
	}
	wg.Wait()
	ctl.progress.Stop()
	return flushErrors(q)
}

func (ctl *Controller) register(runners []Runner) []Runner {
	for i := range runners {
		if ctl.dryRun {
			runners[i].Run = toDryRun(runners[i].Action)
			runners[i].Action = fmt.Sprintf("Dry run: %s", runners[i].Action)
		}
		ctl.progress.AddBar(runners[i].Action, runners[i].Target.GetName())
	}
	return runners
}

func enqueue(q *queue.Queue, items ...interface{}) {
	for _, item := range items {
		if err := q.Put(item); err != nil {
			// This only happens if we are using a Queue after Dispose()
			// has been called on it.
			panic(err)
		}
	}
}

func flushErrors(q *queue.Queue) error {
	if q.Empty() {
		q.Dispose()
		return nil
	}
	length := q.Len()
	items, qErr := q.Get(length)
	if qErr != nil {
		// Get() only returns an error if Dispose() has already
		// been called on the queue.
		panic(qErr)
	}
	q.Dispose()
	return fmt.Errorf("%d fatal error(s) occurred:\n%v", length, items)
}
