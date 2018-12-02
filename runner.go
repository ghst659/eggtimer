// Package hourglass provides functions to time the events reported by a program
// on stdout.
package eggtimer

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"time"
)

// A Clock is a mockable interface providing the necessary methods
// for Run() to get the current time.
type Clock interface {
	Now() time.Time
}

// A RealClock is an implementation of Clock that uses the actual
// clock implemented by the standard library "time" package.
type RealClock struct {}

// Now returns the current local time.
func (r RealClock) Now() time.Time {
	return time.Now()
}

// An Event represents an event reported as a string output on either stdout or stderr
// along with the time after the start of execution, that the event was seen.
type Event struct {
	// The duration between the run start time, and the time that the event was seen.
	When time.Duration
	// The string event seen on stdout or stderr.
	What string
}

// A Runner runs an exec.Cmd, producing an Event on a channel
// for each line of stderr and stdout output.
type Runner struct {
	clock Clock
}

// NewRunner returns a pointer to a new Runner.
func NewRunner(c Clock) *Runner {
	return &Runner{clock: c}
}

// relay is a helper function to turn strings on stdout or stderr into a stream
// of Events into a channel.
func (r *Runner) relay(wg *sync.WaitGroup, p io.ReadCloser, start time.Time, s chan<- Event) error {
	defer wg.Done()
	scanner := bufio.NewScanner(p)
	for scanner.Scan() {
		line := scanner.Text()
		ahora := r.clock.Now()
		delta := ahora.Sub(start)
		s <- Event{
			When: delta,
			What: line,
		}
	}
	return scanner.Err()
}

// Run executes the given exec.Cmd, and produces Events on the given eventStream.
func (r *Runner) Run(c *exec.Cmd, eventStream chan<- Event) {
	defer close(eventStream)
	start := r.clock.Now()
	stdout, err := c.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return
	}
	if err = c.Start(); err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go r.relay(&wg, stdout, start, eventStream)
	go r.relay(&wg, stderr, start, eventStream)
	wg.Wait()
}
