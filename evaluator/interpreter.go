package evaluator

import (
    "errors"
    "fmt"

    "github.com/npillmayer/gorgo/runtime"
    "github.com/npillmayer/gorgo/terex"
    "github.com/npillmayer/pmmp"
)

// ErrNoProgramToExecute flags an empty input program
var ErrNoProgramToExecute error = errors.New("no program to execute")

// Interpreter interprets PMMP programs.
type Interpreter struct {
    evaluator *Evaluator   // expression evaluator and interpreter runtime
    ast       *terex.GCons // program code to execute
    env       *terex.Environment
    thread0   *Thread // initial execution 'thread'
}

// NewInterpreter creates a new interpreter for the PMMP language.
func NewInterpreter() *Interpreter {
    intp := &Interpreter{
        evaluator: NewEvaluator(),
    }
    return intp
}

// Thread is an entity for fetch-decode-excuting AST elements.
type Thread struct {
    PC       *terex.GCons                // program counter
    IR       instruction                 // instruction register
    mem      *runtime.DynamicMemoryFrame // TODO ?
    args     *terex.GCons                // input args
    argsCh   chan *terex.GCons           // channel for input args
    resultCh chan *terex.GCons           // channel for result
    intp     *Interpreter                // life-line to interpreter
    envLocal *terex.Environment          // thread local data
}

// Fork splits an FDE-thread and starts the child thread, processing pc.
func (th Thread) Fork(pc *terex.GCons) *Thread {
    sc := th.intp.evaluator.ScopeTree.PushNewScope(pc.Car.String())
    t := &Thread{
        PC:       pc,
        mem:      runtime.NewDynamicMemoryFrame(pc.Car.String(), sc),
        argsCh:   make(chan *terex.GCons),
        resultCh: make(chan *terex.GCons),
        intp:     th.intp,
    }
    createThreadLocalEnv(t)
    go t.run()
    return t
}

// Args is a channel to pass arguments to an FDE-thread.
func (th *Thread) Args() chan<- *terex.GCons {
    return th.argsCh
}

// Result is a channel to receive a result from an FDE-thread.
func (th *Thread) Result() <-chan *terex.GCons {
    return th.resultCh
}

// Run starts fetch-decode-execute of a thread.
func (th *Thread) run() {
    tracer().Debugf("thread starts at %s", terex.Elem(th.PC.Car))
    th.args = <-th.argsCh // synchronous wait for args, even if none
    tracer().Debugf("thread received arguments %s", terex.Elem(th.args))
    tracer().Debugf("PC = %v", terex.Elem(th.PC))
    result := terex.Elem(nil)
    if th.PC != nil && !isEOF(th.PC) {
        tracer().Debugf("fetch, decode, execute loop")
        result = th.FetchDecodeExecute(terex.Elem(th.PC))
        // if result.AsAtom().Type() == terex.ErrorType {
        //     break
        // } // TODO other than tree-driven FDE ?
        //th.PC = th.PC.Cdr
    }
    th.resultCh <- result.AsList()
}

// FetchDecodeExecute processes a fragment of the AST.
func (th *Thread) FetchDecodeExecute(e terex.Element) terex.Element {
    var err error
    if e.Type() == terex.NumType { // TODO pack this into package pmmp
        return terex.Elem(pmmp.FromFloat(e.AsAtom().Data.(float64)))
    } else if e.Type() == terex.StringType {
        // TODO string Value
    }
    th.IR, err = th.intp.fetch(e.AsList()) // fetch
    if err != nil {
        tracer().Errorf("FDE error: " + err.Error())
        th.intp.env.Error(err)
        return terex.Elem(terex.ErrorAtom(err.Error()))
    }
    // decode
    tracer().Debugf("decode instruction %v", th.IR)
    result := th.IR(e, th.envLocal) // execute
    return result
}

// Start expects a list (AST #eof)
func (intp *Interpreter) Start(program *terex.GCons, env *terex.Environment) (*terex.GCons, error) {
    intp.ast = program
    if env == nil {
        terex.InitGlobalEnvironment()
        env = terex.GlobalEnvironment
    }
    intp.env = env
    env.Def("$Evaluator", terex.Elem(intp.evaluator))
    eof := intp.ast.Last()
    if eof == intp.ast {
        tracer().Errorf("empty program?")
        return nil, ErrNoProgramToExecute
    }
    //intp.PC = intp.ast
    intp.thread0 = Thread{intp: intp}.Fork(terex.Elem(intp.ast.Tee()).AsList())
    intp.thread0.Args() <- nil
    r := <-intp.thread0.Result()
    tracer().Infof("statement execution returned %v", terex.Elem(r))
    return r, nil
}

// Fetch the instruction belonging to operator op.
// We do not call the operator directly, but rather search for an operator-symbol
// in the current environment and fetch its 'Call' method.
//
// This indirection allows the AST to be built out of empty operator tags,
// which do nothing. Functionality can be changed by providing different
// operators in the environment (or completely swapping environments with
// different operator sets pre-loaded).
//
// Will return a NOP if operator is not found in environment.
//
func (intp *Interpreter) fetch(astNode *terex.GCons) (instruction, error) {
    e := terex.Elem(astNode.Car)
    if !e.IsAtom() || e.Type() != terex.OperatorType {
        tracer().Errorf("fetch saw non-operator")
        return nop, fmt.Errorf("op fetch saw: %v", astNode.Car)
    }
    op := e.First().AsAtom().Data.(pmmp.TokenOperator)
    opname := op.Token().Name
    opsym := intp.env.FindSymbol(opname, true)
    if opsym == nil {
        tracer().Errorf("Cannot find operation %s", opname)
        tracer().Debugf(intp.env.Dump())
        return nop, fmt.Errorf("operator not found in env: %v", opname)
    }
    operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
    if !ok {
        tracer().Errorf("Cannot call operation %s", opname)
        return nop, fmt.Errorf("Cannot call operation %s", opname)
    }
    tracer().Debugf("fetch of operator %s", operator)
    return operator.Call, nil
}

// GetEvaluator resolves an Evaluator from an environment. This is to be
// used by operators, which will have been passed an environment which
// includes a symbol for the calling interpreter's evaluator.
//
func GetEvaluator(env *terex.Environment) *Evaluator {
    esym := env.FindSymbol("$Evaluator", true)
    if esym == nil {
        panic("no evaluator present")
    }
    return esym.Value.AsAtom().Data.(*Evaluator)
}

// GetThread resolves a FDE-thread from an environment. This is to be
// used by operators, which will have been passed an environment which
// includes a symbol for the calling interpreter's thread.
//
func GetThread(env *terex.Environment) *Thread {
    tsym := env.FindSymbol("$Thread", true)
    if tsym == nil {
        panic("no thread present")
    }
    return tsym.Value.AsAtom().Data.(*Thread)
}

func createThreadLocalEnv(th *Thread) {
    th.envLocal = terex.NewEnvironment("thread-local", th.intp.env)
    th.envLocal.Def("$Thread", terex.Elem(th))
}

func isEOF(pc *terex.GCons) bool {
    return false // TODO
}

type instruction func(terex.Element, *terex.Environment) terex.Element

func nop(terex.Element, *terex.Environment) terex.Element {
    return terex.Elem(nil)
}
