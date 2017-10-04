package runner

import (
	"fmt"
	"sync"
	"time"

	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/errorqueue"
	progressbars "github.com/benjdewan/pachelbel/progress"
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

// NewController creates a new Controller object
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
// their 'run' function in parallel.
func (ctl *Controller) Run(runners []Runner) error {
	runners = ctl.register(runners)

	var wg sync.WaitGroup
	wg.Add(len(runners))
	q := errorqueue.New()
	ctl.progress.Start()

	for _, runner := range runners {
		go func(r Runner) {
			if err := r.Run(ctl.cxn, r.Target); err != nil {
				ctl.progress.Error(r.Target.GetName())
				q.Enqueue(err)
			} else {
				ctl.progress.Done(r.Target.GetName())
			}
			wg.Done()
		}(runner)
	}
	wg.Wait()
	ctl.progress.Stop()
	return q.Flush()
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
