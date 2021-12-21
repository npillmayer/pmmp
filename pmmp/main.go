// Package pmmp is a poor man's version of MetaFont/MetaPost.
//
// License
//
// Governed by a 3-Clause BSD license. License file may be found in the root
// folder of this module.
//
// Copyright © 2021 Norbert Pillmayer <norbert@pillmayer.com>
//
package main

import (
	"context"
	"os"
	"os/signal"

	"gioui.org/app"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/pmmp/cli"
)

func main() {
	var stop context.CancelFunc
	pmmp.SignalContext, stop = signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// start the CLI in a goroutine, as the main thread will be blocked by the GUI
	go func() {
		cli.Execute()
	}()

	// app.Main will potentially move the focus away from the shell window
	// therefore we guard it until a GUI window is to be opened.
	// CLI commands will have to call os.Exit() to terminate the application.
	pmmp.ConditionGuiStarted.Wait() // we guard app.Main()
	app.Main()
}
