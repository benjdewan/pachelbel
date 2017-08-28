// Copyright © 2017 ben dewan <benj.dewan@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package progress

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressBars is the primary interface into the package.
type ProgressBars struct {
	// Writer can be set to direct where progress bars are printed
	// os.Stdout is used by default
	Writer io.Writer

	// Width specifies the overall width of the progress bars display.
	// 80 is used by default
	Width int

	// RefreshRate controls the frequency of updates being pushed to the
	// screen.
	// 1 second is the default
	RefreshRate time.Duration

	// internal fields
	bars     []*progressBar
	started  bool
	stopChan chan struct{}
	lock     *sync.RWMutex
}

const (
	// ActionUpdate indicates that a deployment is being updated
	ActionUpdate = "Updating"
	// ActionCreate indicates that a deployment is being created
	ActionCreate = "Creating"

	stateRunning  = "running"
	stateDone     = "done"
	stateFailed   = "failed"
	stateFinished = "finished"
)

const bar = "░"

type progressBar struct {
	action string
	name   string
	state  string
}

// New provisions a new ProgressBars struct with default values
func New() *ProgressBars {
	return &ProgressBars{
		Writer:      os.Stdout,
		Width:       80,
		RefreshRate: time.Second,
		bars:        [](*progressBar){},
		stopChan:    make(chan struct{}, 1),
		lock:        &sync.RWMutex{},
	}
}

// AddBar adds a progress bar to the ProgressBars struct. The action and name
// are used to print status information when the progress bars are started,
// Because this package is designed to work in non-tty settings where
// re-painting screens doesn't work we cannot dynamically add new progress
// bars, so this method panics if the ProgressBars is currently running when
// invoked.
func (p *ProgressBars) AddBar(action, name string) *ProgressBars {
	if p.started {
		panic("Progress bars cannot be added while running")
	}
	p.bars = append(p.bars, &progressBar{
		action: action,
		name:   name,
		state:  stateRunning,
	})
	return p
}

// Start prints a state header defining the progress bars and then begins
// printing the bars themselves. It will continue printing until every bar
// is finished or Stop() is called.
func (p *ProgressBars) Start() {
	p.started = true
	p.printHeader()

	go func() {
		for {
			select {
			case <-p.stopChan:
				return
			default:
				p.draw()
				time.Sleep(p.RefreshRate)
			}
		}
	}()

}

// Done terminates a single progress bar by name in a successful state
func (p *ProgressBars) Done(barName string) {
	p.changeState(barName, stateDone)
}

// Done terminates a single progress bar by name in a failure state
func (p *ProgressBars) Error(barName string) {
	p.changeState(barName, stateFailed)
}

// Stop ends the printing of all progress bars.
func (p *ProgressBars) Stop() {
	p.stopChan <- struct{}{}
	p.started = false
}

func (p *ProgressBars) changeState(name, state string) {
	for _, bar := range p.bars {
		if bar.name != name {
			continue
		}
		p.lock.Lock()
		bar.state = state
		p.lock.Unlock()
		return
	}
}

func (p *ProgressBars) draw() {
	width := p.barWidth()
	p.lock.Lock()
	statuses := []string{}
	for _, bar := range p.bars {
		switch bar.state {
		case stateRunning:
			statuses = append(statuses, strings.Repeat("░", width))
		case stateDone:
			statuses = append(statuses, center("DONE", width))
			bar.state = stateFinished
		case stateFailed:
			statuses = append(statuses, center("ERROR", width))
			bar.state = stateFinished
		case stateFinished:
			statuses = append(statuses, strings.Repeat(" ", width))
		default:
			log.Panicf("Unknown progress bar state: %s", bar.state)
		}
	}
	p.lock.Unlock()

	line := strings.Join(statuses, " ")
	if len(strings.Trim(line, " ")) == 0 { // everything has finished. Stop drawing
		p.Stop()
		return
	}

	fprintln(p.Writer, strings.Join(statuses, " "))
}

func (p *ProgressBars) printHeader() {
	width := p.barWidth()
	barHeaders := []string{}
	if maxWidth(p.bars) < width {
		barHeaders = p.completeHeaders(width)
	} else if maxNameWidth(p.bars) < width {
		barHeaders = p.namedHeaders(width)
	} else {
		barHeaders = p.numberedHeaders(width)
	}
	fprintln(p.Writer, strings.Join(barHeaders, " "))
}

func maxWidth(bars []*progressBar) int {
	max := 0
	for _, bar := range bars {
		str := fmt.Sprintf("%s '%s'", bar.action, bar.name)
		if len(str) > max {
			max = len(str)
		}
	}
	return max
}

func maxNameWidth(bars []*progressBar) int {
	max := 0
	for _, bar := range bars {
		if len(bar.name) > max {
			max = len(bar.name)
		}
	}
	return max
}

func (p *ProgressBars) completeHeaders(barWidth int) []string {
	barHeaders := []string{}
	for _, bar := range p.bars {
		barHeaders = append(barHeaders,
			center(fmt.Sprintf("%s '%s'", bar.action, bar.name),
				barWidth))
	}
	return barHeaders
}

func (p *ProgressBars) namedHeaders(barWidth int) []string {
	barHeaders := []string{}
	for _, bar := range p.bars {
		barHeaders = append(barHeaders, center(bar.name, barWidth))
	}
	return barHeaders
}

func (p *ProgressBars) numberedHeaders(barWidth int) []string {
	barHeaders := []string{}
	for i, bar := range p.bars {
		num := fmt.Sprintf("%02d", i)
		fprintf(p.Writer, "%02d - %s '%s'\n", i, bar.action, bar.name)
		barHeaders = append(barHeaders, center(num, barWidth))
	}
	return barHeaders
}

func (p *ProgressBars) barWidth() int {
	return (p.Width - len(p.bars)) / len(p.bars)
}

func center(str string, width int) string {
	if len(str) > width {
		return strings.Join(strings.SplitAfter(str, "")[:width], "")
	}
	div := (width - len(str)) / 2

	return strings.Repeat(" ", div) + str + strings.Repeat(" ", div)
}

func fprintln(w io.Writer, a ...interface{}) {
	if _, err := fmt.Fprintln(w, a...); err != nil {
		panic(err)
	}
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		panic(err)
	}
}
