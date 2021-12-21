// Package termui provides objects and methods for interactive UI in terminal windows.
//
// License
//
// Governed by a 3-Clause BSD license. License file may be found in the root
// folder of this module.
//
// Copyright © 2021 Norbert Pillmayer <norbert@pillmayer.com>
//
//
package termui

import (
	"encoding/json"
	"fmt"
	"image"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/npillmayer/schuko/tracing"
)

// trace traces with key 'pmmp.cli'.
func trace() tracing.Trace {
	return tracing.Select("pmmp.cli")
}

type Formatter interface {
	Format(interface{}, io.Writer) (bool, error)
}

type DefaultFormatter struct{}

func (df DefaultFormatter) Format(item interface{}, w io.Writer) (bool, error) {
	switch t := item.(type) {
	case string:
		w.Write([]byte("▶ "))
		if _, err := w.Write([]byte(t)); err != nil {
			return false, err
		}
		w.Write([]byte{'\n'})
		return true, nil
	case map[string]interface{}:
		jsn, err := json.MarshalIndent(t, "  ", "    ")
		if err != nil {
			return false, nil
		}
		w.Write([]byte("▶ Hierarchical object: "))
		w.Write(jsn)
		w.Write([]byte{'\n'})
		return true, nil
	case image.Image:
		dims := t.Bounds()
		w.Write([]byte("▶ "))
		w.Write([]byte(fmt.Sprintf("image of size %d x %d\n", dims.Dx(), dims.Dy())))
		return true, nil
	case table.Writer:
		if t == nil {
			w.Write([]byte("▶ (empty table)\n"))
		} else {
			w.Write([]byte(t.Render()))
			w.Write([]byte{'\n'})
		}
		return true, nil
	default:
		w.Write([]byte("▶ "))
		w.Write([]byte(fmt.Sprintf("object of type %T\n", t)))
		return true, nil
	}
	//return false, nil
}
