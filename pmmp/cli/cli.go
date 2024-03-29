// Package cli implements the pmmp command line interface.
//
// License
//
// Governed by a 3-Clause BSD license. License file may be found in the root
// folder of this module.
//
// Copyright © 2021 Norbert Pillmayer <norbert@pillmayer.com>
//
package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/pmmp/ui/termui"
	"github.com/npillmayer/schuko/tracing"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pmmp",
	Short: "A poor man's version of MetaFont/MetaPost",
	Long: `Welcome to PMMP V0.1 (experimental)

PMMP lets you interpret programs written in MetaFont/MetaPost format.

PMMP is able to run in interactive mode or execute one or more commands in
batch-mode.  If run in interactive mode, it will prompt for user input in a
terminal REPL and is able to additionally show certain kinds of information
in a GUI window.

`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: runPmmpCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called exactly once by pmmp.main().
func Execute() {
	//rootCmd.AddCommand(fonts.Command())
	if rootCmd.Execute() != nil {
		pmmp.Exit(2)
	}
}

func init() {
	cobra.OnInitialize(loadConfig)
	// persistent flags which will be global for the application
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "Force run in interactive mode")
	rootCmd.PersistentFlags().String("logfile", "stderr", "URL of log output location")
}

// TODO if -c <cmd> flag is given:
// if -i flag present:
// - execute command with REPL
// - send global gui app a cancel signal
// else:
// - execute command without REPL
func runPmmpCmd(cmd *cobra.Command, args []string) {
	runPmmpCmdIntpr(cmd, args)
}

func runPmmpCmdIntpr(cmd *cobra.Command, args []string) {
	tracing.Infof("pmmp interpreter called")
	fcmd := &pmmpCmdIntpr{}
	fcmd.BaseREPL = termui.NewBaseREPL("pmmp", "0.1 experimental")
	fcmd.Interpreter = fcmd
	fcmd.Helper = func(w io.Writer) {
		io.WriteString(w, `
tylot-fonts will interpret the following statements:

  find-font <pattern> | <descriptor> -> find-font : locate and load a font
  inspect <font>                                  : store font for inspection
  font.info                                       : print informations about the inspected font
  glyph.info <glyphindex> | <codepoint>           : print information about a glyph
  show                                            : display visual information

`)
	}
	//stdout, stderr := fcmd.Outputs()
	//fcmd.ElvishInterpreter = termui.NewElvishInterpreter(stdout, stderr)
	// TODO
	//fcmd.ElvishInterpreter = nil
	fcmd.addInterpreterStatements()
	fcmd.Prompt(true)
}

type pmmpCmdIntpr struct {
	*termui.BaseREPL
	mpPipe io.WriteCloser
	//*termui.ElvishInterpreter
}

func (fcmd *pmmpCmdIntpr) InterpretCommand(command string) {
	//tracer().Debugf("font interpreter: %q\n", command)
	command = strings.Trim(command, " \t\x00")
	//err := fcmd.Eval(command, Formatter{})
	err := fmt.Errorf("command not found: %q", command)
	if err != nil {
		_, stderr := fcmd.Outputs()
		fmt.Fprintf(stderr, "interpreter error: %s\n", err.Error())
	}
}

func (fcmd *pmmpCmdIntpr) addInterpreterStatements() {
	// fcmd.AddBuiltinCommands(map[string]interface{}{
	// 	"show": show,
	// })

	// TODO
	// pipe := termui.NewBytePipe()
	//grammar.Parse(bufio.NewReader(pipe.Reader()))
}
