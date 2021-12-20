package grammar

import (
	"bytes"
	"io"
	"unicode/utf8"
)

type runeStream struct {
	isEof      bool
	next       rune
	rlen       int
	start, end uint64 // as bytes index
	reader     io.RuneReader
	writer     bytes.Buffer
}

func (rs runeStream) OutputString() string {
	return rs.writer.String()
}

func (rs *runeStream) ResetOutput() {
	rs.writer.Reset()
}

func (rs runeStream) Span() uint64 {
	return rs.end - rs.start
}

func (rs *runeStream) lookahead() (r rune, err error) {
	if rs.isEof {
		return 0, io.EOF
	}
	if rs.next != 0 {
		r = rs.next
		tracer().Debugf("read LA %#U", r)
		rs.next = 0
		return
	}
	var sz int
	r, sz, err = rs.reader.ReadRune()
	rs.next = r
	rs.rlen += sz
	if err == io.EOF {
		tracer().Debugf("EOF for MetaPost input")
		rs.isEof = true
		r = utf8.RuneError
		return
	} else if err != nil {
		return 0, err
	}
	tracer().Debugf("read rune %#U", r)
	return
}

func (rs *runeStream) match(r rune) {
	if rs.isEof {
		return
	}
	//l.lexeme.WriteRune(r)
	if rs.next != 0 {
		rs.next = 0
		return
	}
}
