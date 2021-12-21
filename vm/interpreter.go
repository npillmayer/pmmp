package vm

import (
	"errors"

	"github.com/npillmayer/pmmp/sframe"
)

// ErrNoProgramToExecute flags an empty input program
var ErrNoProgramToExecute error = errors.New("no program to execute")

type VM struct {
	threads   []Thread
	heap      sframe.ScopeFrameTree
	input     chan Op // op input stream
	lastError error
}

func (vm *VM) SetError(err error) {
	vm.lastError = err
}

// Thread is an entity for fetch-decode-excuting AST elements.
//
// TODO add a context
type Thread struct {
	//PC int // program counter
	//IR    instruction               // instruction register
	stack *ExprStack                // thread local stack
	heap  *sframe.DynamicScopeFrame // private and global memory frames
	regs  *RegisterSet              // registers to store op arguments in
	code  <-chan Op                 // where to read instructions from
	vm    *VM                       // life-line back to VM
}

// Fork splits an FDE-thread and starts the child thread, processing pc.
func (vm *VM) Fork(code <-chan Op, globalMem *sframe.DynamicScopeFrame) *Thread {
	frame := vm.heap.PushNewFrame(0)
	vm.threads = append(vm.threads, Thread{
		code:  code,
		stack: NewExprStack(),
		heap:  frame,
		vm:    vm,
	})
	thread := &vm.threads[len(vm.threads)-1]
	go thread.run()
	return thread
}

// Run starts fetch-decode-execute of a thread.
func (thread *Thread) run() {
	tracer().Debugf("thread starts fetch, decode, execute loop")
	var err error
	for op := range thread.code {
		thread.regs.DecodeArg(op)
		err = thread.Execute(op)
		if err != nil {
			tracer().Errorf("thread error executing %02x", op.opcode)
			thread.vm.SetError(err)
		}
	}
}

func (thread *Thread) Execute(op Op) error {
	tracer().Debugf("thread executing %2x %v", op.opcode, op.arg)
	// TODO command pattern
	// first draft with switch
	switch op.opcode {
	case IConst:
		tracer().Debugf("putting I constant %d onto the stack", thread.regs.I)
	case FConst:
		tracer().Debugf("putting F constant %.4f onto the stack", thread.regs.F)
	}
	return nil
}
