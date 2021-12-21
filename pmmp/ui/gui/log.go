package gui

import (
	"fmt"
	"strings"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// --- Example for small auto-pilot view windows ------------------
//
// Copy this idea for windows displaying elv.sh object output

// nouseLetters displays a clickable list of text items that open a new window.
type nouseLetters struct {
	win *Window
	log *Log

	items []*nouseLetterListItem
	list  widget.List
}

// nouseNewLetters creates a new letters view with the provided log.
func nouseNewLetters(log *Log) *nouseLetters {
	view := &nouseLetters{
		log:  log,
		list: widget.List{List: layout.List{Axis: layout.Vertical}},
	}
	for text := 'a'; text <= 'z'; text++ {
		view.items = append(view.items, &nouseLetterListItem{Text: string(text)})
	}
	return view
}

// Run implements Window.Run method.
func (v *nouseLetters) Run(w *Window) error {
	v.win = w
	return WidgetView(v.Layout).Run(w)
}

// Layout handles drawing the letters view.
func (v *nouseLetters) Layout(gtx layout.Context) layout.Dimensions {
	th := v.win.App.Theme
	return material.List(th, &v.list).Layout(gtx, len(v.items), func(gtx layout.Context, index int) layout.Dimensions {
		item := v.items[index]
		for item.Click.Clicked() {
			v.log.Printf("opening %s view", item.Text)

			bigText := material.H1(th, item.Text)
			size := bigText.TextSize
			size.V *= 2
			v.win.App.NewWindow(item.Text,
				WidgetView(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, bigText.Layout)
				}),
				app.Size(size, size),
			)
		}
		return material.Button(th, &item.Click, item.Text).Layout(gtx)
	})
}

type nouseLetterListItem struct {
	Text  string
	Click widget.Clickable
}

// --- "Console log" window --------------------------------------------------

// Log shows a list of strings.
type Log struct {
	addLine chan string
	lines   []string

	list widget.List
}

// newLog creates a new log view.
func newLog() *Log {
	return &Log{
		addLine: make(chan string, 100),
		list:    widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

// Printf adds a new line to the log.
func (log *Log) Printf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	s = strings.TrimSuffix(s, "\n")
	S := strings.Split(s, "\n")
	for _, line := range S {
		select { // ensure that this logging does not block
		case log.addLine <- line:
		default:
		}
	}
}

// Run handles window loop for the log.
func (log *Log) Run(w *Window) error {
	var ops op.Ops
	applicationClose := w.App.Context.Done()
	for {
		select {
		case <-applicationClose:
			return nil
		// listen to new lines from Log.Printf and add them to widget's lines
		case line := <-log.addLine:
			log.lines = append(log.lines, line)
			w.Invalidate()
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				GlobalGui().CloseLog()
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				log.Layout(w, w.App.Theme, gtx)
				e.Frame(gtx.Ops)
			}
		}
	}
}

func (log *Log) Layout(w *Window, th *material.Theme, gtx layout.Context) {
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &log.list).Layout(gtx, len(log.lines), func(gtx layout.Context, i int) layout.Dimensions {
				return material.Body1(th, log.lines[i]).Layout(gtx)
			})
		}),
	)

}
