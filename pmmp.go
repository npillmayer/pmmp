package pmmp

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/knadh/koanf"
)

// Configuration holds global configuration values. We use koanf.
var Configuration *koanf.Koanf

// Tracefile is the file we write our log output, if not nil.
var Tracefile io.WriteCloser

// SignalContext is a global context for terminating the application by an interrupt
// signal.
var SignalContext context.Context

// Exit exits the application. It gracefully shuts down all resources.
func Exit(errcode int) {
	if Tracefile != nil {
		Tracefile.Close()
	}
	os.Exit(errcode)
}

// ConditionGuiStarted is a condition variable for waiting/announcing that a window
// has been opened and therefore the GUI subsystem has been ramped up.
//
// Wait for this condition with `ConditionGuiStarted.Wait()` and announce it with
// `ConditionGuiStarted.Broadcast()`.
//
var ConditionGuiStarted = newCondition("guiStarted")

type condition struct {
	condition *sync.Cond // we use a condition variable
	mu        sync.Mutex // guards the condition
	variable  bool       // the condition to guard
	name      string     // identifies this condition
}

func newCondition(name string) *condition {
	c := &condition{
		mu:   sync.Mutex{},
		name: name,
	}
	c.condition = sync.NewCond(&c.mu)
	return c
}

func (c *condition) Wait() {
	c.condition.L.Lock()
	for !c.variable {
		c.condition.Wait()
	}
	c.condition.L.Unlock()
}

func (c *condition) Broadcast() {
	c.condition.L.Lock()
	c.variable = true
	c.condition.Broadcast()
	c.condition.L.Unlock()
}
