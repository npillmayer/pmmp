package grammar

import (
	"bufio"
	"io"
	"strings"

	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/pmmp/sframe"
)

type nestedReader struct {
	reader io.RuneReader
	parent *nestedReader
}

func (nr *nestedReader) ReadRune() (r rune, size int, err error) {
	for {
		r, size, err = nr.reader.ReadRune()
		if err != nil && err != io.EOF {
			return
		}
		if err == io.EOF && nr.parent != nil {
			nr = nr.parent
			continue
		}
		return
	}
}

func (nr *nestedReader) Push(rr io.RuneReader) *nestedReader {
	subnr := &nestedReader{
		reader: rr,
		parent: nr,
	}
	return subnr
}

func (nr *nestedReader) PushMacro(v sframe.Variable, env *terex.Environment) *nestedReader {
	macro := v.(sframe.Macro)
	bound := make([]sframe.Variable, 0, len(macro.ArgsList))
	for i, arg := range macro.ArgsList {
		bound[i] = bind(arg, env)
	}
	macroFrame := sframe.GlobalFrameStack.PushNewFrame(macro.Declaration())
	for i, variable := range bound {
		macroFrame.Env().Def(macro.ArgsList[i].Name(), terex.Elem(variable))
	}
	return &nestedReader{
		reader: bufio.NewReader(strings.NewReader(macro.ReplacementText())),
		parent: nr,
	}
}

func bind(sym sframe.TagDecl, env *terex.Environment) sframe.Variable {
	return sframe.Numeric{}
}
