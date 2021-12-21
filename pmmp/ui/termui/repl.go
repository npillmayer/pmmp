package termui

// Utilities for interactive command line interfaces.

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	prtxt "github.com/jedib0t/go-pretty/v6/text"
	"github.com/npillmayer/pmmp"
)

// Some global defaults
var welcomeMessage = "Welcome to %s [V%s]"
var stdprompt = prtxt.FgGreen.Sprint("%s> ")
var editmode string = "emacs"

// BaseREPL is a base type to instantiate a REPL interpreter.
// Concrete REPL implementations will usually use this as a base type.
type BaseREPL struct {
	Interpreter REPLCommandInterpreter // the interpreter this REPL runs for
	Helper      func(io.Writer)        // print out help information
	readline    *readline.Instance
	toolname    string
	version     string
}

// NewBaseREPL create a new REPL base object intialized for an interpreter tool
// and a given version.
func NewBaseREPL(toolname, version string) *BaseREPL {
	repl := &BaseREPL{
		readline: newReadline(toolname, version),
		toolname: toolname,
		version:  version,
	}
	return repl
}

// REPLCommandInterpreter is an interface all interpreters have to implement.
// It is the workhorse doing interpretation of interactive
// commands.
//
// The REPL will delegate interpreting command strings (i.e. those which do not represent
// internal administrative commands) to the interpreter.
type REPLCommandInterpreter interface {
	InterpretCommand(string)
}

// Create a readline instance.
func newReadline(toolname, version string) *readline.Instance {
	histfile := fmt.Sprintf("%s/%s-repl-history.tmp", os.TempDir(), toolname)
	prompt := fmt.Sprintf(stdprompt, toolname)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              prompt,
		HistoryFile:         histfile,
		AutoComplete:        replCompleter,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterReplInput,
	})
	if err != nil {
		panic(err)
	}
	return rl
}

// displayCommands prints a help message with available commands
// We support some internal interactive sub-commands (not part of the interpreter).
func (repl *BaseREPL) displayCommands(out io.Writer) {
	io.WriteString(out, fmt.Sprintf(welcomeMessage, repl.toolname, repl.version))
	io.WriteString(out, "\n\nThe following commands are available:\n\n")
	io.WriteString(out, "  help               : print this message\n")
	io.WriteString(out, "  bye                : quit application\n")
	io.WriteString(out, "  mode [mode]        : display or set current editing mode\n")
	io.WriteString(out, "  setprompt [prompt] : set current prompt [to default],\n")
}

// Completer-tree for interactive frames sub-commands
var replCompleter = readline.NewPrefixCompleter(
	readline.PcItem("help"),
	readline.PcItem("bye"),
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("setprompt"),
)

// Outputs returns stdout and stderr of this REPL.
func (repl *BaseREPL) Outputs() (io.Writer, io.Writer) {
	return repl.readline.Stdout(), repl.readline.Stderr()
}

// Prompt enters a REPL and executes commands.
// Commands are either internal administrative (setprompt, help, etc.)
// or interpreted statements.
func (repl *BaseREPL) Prompt(exitOnBye bool) {
	defer repl.readline.Close()
	io.WriteString(repl.readline.Stderr(),
		fmt.Sprintf(welcomeMessage, repl.toolname, repl.version))
	if !strings.HasSuffix(welcomeMessage, "\n") {
		repl.readline.Stderr().Write([]byte{'\n'})
	}
	for {
		line, err := repl.readline.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		words := strings.Fields(line)
		command := ""
		if len(words) > 0 {
			command = words[0]
		}
		if doExit := repl.executeCommand(command, words, line); doExit {
			break
		}
	}
	if exitOnBye {
		pmmp.Exit(0)
	}
}

// Central dispatcher function to execute internal REPL commands or interpreter
// statements. It receives the command (i.e. the first word of the line),
// a list of words (args) including the command, and the complete line of text.
// If it returns true, the REPL should terminate.
func (repl *BaseREPL) executeCommand(cmd string, args []string, line string) bool {
	switch {
	case cmd == "":
		// do nothing
	case cmd == "help":
		repl.displayCommands(repl.readline.Stderr())
		if repl.Helper != nil {
			repl.Helper(repl.readline.Stderr())
		}
	case cmd == "bye":
		println("> goodbye!")
		return true
	case cmd == "mode":
		if len(args) > 1 {
			switch args[1] {
			case "vi":
				repl.readline.SetVimMode(true)
				editmode = "vi"
				return false
			case "emacs":
				repl.readline.SetVimMode(false)
				editmode = "emacs"
				return false
			}
		}
		io.WriteString(repl.readline.Stderr(),
			fmt.Sprintf("> current input mode: %s\n", editmode))
	case cmd == "setprompt":
		var prmpt string
		if len(line) <= 10 {
			prmpt = fmt.Sprintf(stdprompt, repl.toolname)
		} else {
			prmpt = line[10:] + " "
		}
		repl.readline.SetPrompt(prmpt)
	case cmd == "":
	default:
		trace().Debugf("call interpreter on: '%s'", line)
		repl.interpret(line)
	}
	return false // do not exit
}

// interpret calls the interpreter, sending a statement.
func (repl *BaseREPL) interpret(line string) {
	if repl.Interpreter == nil {
		return
	}
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		io.WriteString(repl.readline.Stderr(), "> error executing statement!\n")
	// 		io.WriteString(repl.readline.Stderr(), fmt.Sprintf("> %v\n", r)) // TODO: get ERROR and print
	// 	}
	// }()
	repl.Interpreter.InterpretCommand(line)
}

// Input filter for REPL. Blocks ctrl-z.
func filterReplInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
