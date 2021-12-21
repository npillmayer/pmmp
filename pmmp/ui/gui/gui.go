package gui

import (
	"context"
	"image"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/npillmayer/pmmp"

	"gioui.org/font/gofont"
)

// Example for a GIO event management loop
/*
func Loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			l := material.H1(th, "Hello, Gio")
			maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
			l.Color = maroon
			l.Alignment = text.Middle
			l.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
*/

// --- New stuff -------------------------------------------------------------

var guiApplication *GuiApplication

var guiAppStart sync.Once

func GlobalGui() *GuiApplication {
	guiAppStart.Do(func() {
		guiApplication = NewApplication(pmmp.SignalContext)
		//go func() {
		pmmp.ConditionGuiStarted.Broadcast()
		// defer guiApplication.mx.Unlock()
		// guiApplication.mx.Lock()
		// guiApplication.log = newLog()
		// now := time.Now()
		// guiApplication.log.Printf("[Log output startet at %s]", now.Format(time.UnixDate))
		// guiApplication.NewWindow("Log", guiApplication.log)
		//
		//guiApplication.Wait()
		//}()
	})
	return guiApplication
}

// GuiApplication keeps track of all the windows and global state.
type GuiApplication struct {
	Context  context.Context // used to broadcast application shutdown
	Shutdown func()          // shut down all windows
	Theme    *material.Theme // the application wide theme
	active   sync.WaitGroup  // keep track of open windows
	log      *Log            // "console log" window to show text
	mx       sync.Mutex      // guards log
}

func NewApplication(ctx context.Context) *GuiApplication {
	ctx, cancel := context.WithCancel(ctx)
	return &GuiApplication{
		Context:  ctx,
		Shutdown: cancel,
		Theme:    material.NewTheme(gofont.Collection()),
	}
}

// Wait waits for all windows to close.
func (a *GuiApplication) Log() *Log {
	defer a.mx.Unlock()
	a.mx.Lock()
	if a.log == nil {
		a.log = newLog()
		now := time.Now()
		a.log.Printf("[Log output startet at %s]", now.Format(time.UnixDate))
		a.NewWindow("Log", guiApplication.log)
	}
	return a.log
}

func (a *GuiApplication) CloseLog() {
	defer a.mx.Unlock()
	a.mx.Lock()
	a.log = nil
}

func (a *GuiApplication) Wait() {
	a.active.Wait()
}

// NewWindow creates a new tracked window.
func (a *GuiApplication) NewWindow(title string, view View, opts ...app.Option) {
	opts = append(opts, app.Title(title))
	w := &Window{
		App:    a,
		Window: app.NewWindow(opts...),
	}
	a.active.Add(1)
	go func() {
		defer a.active.Done()
		view.Run(w)
	}()
}

// Window holds window state.
type Window struct {
	App *GuiApplication
	*app.Window
}

// View describes .
type View interface {
	// Run handles the window event loop.
	Run(w *Window) error
}

// --- the following is used for situation like the 'letter' app-window --------

// WidgetView allows to use layout.Widget as a view.
type WidgetView func(gtx layout.Context) layout.Dimensions

// Run displays the widget with default handling.
func (view WidgetView) Run(w *Window) error {
	var ops op.Ops

	applicationClose := w.App.Context.Done()
	for {
		select {
		case <-applicationClose:
			return nil
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				view(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}
}

// --- Image view ------------------------------------------------------------

func NewImageView(img image.Image) (WidgetView, []app.Option, error) {
	bounds := img.Bounds()
	sizingOption := app.Size(unit.Dp(float32(bounds.Dx())), unit.Dp(float32(bounds.Dy())))
	return WidgetView(func(gtx layout.Context) layout.Dimensions {
		imgw := widget.Image{
			Src: paint.NewImageOp(img),
			Fit: widget.Contain,
		}
		return imgw.Layout(gtx)
	}), []app.Option{sizingOption}, nil
}
