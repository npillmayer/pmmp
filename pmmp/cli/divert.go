package cli

import (
	"fmt"
	"image"
	"io"

	"gioui.org/app"
	"github.com/npillmayer/pmmp/pmmp/ui/gui"
	"github.com/npillmayer/pmmp/pmmp/ui/termui"
)

func getViewFor(object interface{}) (gui.View, []app.Option, error) {
	switch t := object.(type) {
	case image.Image:
		return gui.NewImageView(t)
	default:
		tracer().Debugf("no view could be decided for type %T", t)
	}
	return nil, nil, nil
}

type Formatter struct {
	termui.DefaultFormatter
}

func (f Formatter) Format(item interface{}, w io.Writer) (bool, error) {
	tracer().Debugf("fonts.Format called for item %T", item)
	switch t := item.(type) {
	case int32:
		w.Write([]byte(fmt.Sprintf("▶ %#U\n", t)))
		return true, nil
		// case glyphDetails:
		// 	tw := glyphDetailsAsTable(t)
		// 	if tw == nil {
		// 		w.Write([]byte(fmt.Sprintf("▶ %d\n", t)))
		// 		return true, nil
		// 	}
		// 	item = tw
	}
	return f.DefaultFormatter.Format(item, w)
}

// --- Property tables for various types -------------------------------------

// func glyphDetailsAsTable(gi glyphDetails) table.Writer {
// 	tw := table.NewWriter()
// 	tw.SetTitle("Glyph %d of %q", gi.gid, currentOTFont.F.Fontname)
// 	tw.AppendRow(table.Row{
// 		"glyph index",
// 		fmt.Sprintf("%d", gi.gid),
// 	})
// 	cp := "–"
// 	if gi.codepoint != 0 {
// 		cp = fmt.Sprintf("%#U", gi.codepoint)
// 	}
// 	tw.AppendRow(table.Row{"code-point", cp})
// 	tw.AppendRow(table.Row{"glyph class", gi.Class})
// 	tw.AppendRow(table.Row{"mark attachment class", gi.MarkAttachClass})
// 	tw.AppendRow(table.Row{"mark glyph set", gi.MarkGlyphSet})
// 	tw.SetStyle(table.StyleLight)
// 	return tw
// }
