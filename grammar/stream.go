package grammar

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/npillmayer/gorgo"
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
	rs.start = rs.end
}

func (rs runeStream) Span() gorgo.Span {
	return gorgo.Span{rs.start, rs.end}
}

func (rs *runeStream) lookahead() (r rune, err error) {
	if rs.isEof {
		return utf8.RuneError, io.EOF
	}
	if rs.next != 0 {
		r = rs.next
		tracer().Debugf("read LA %#U", r)
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
	tracer().Debugf("match %#U", r)
	if r == utf8.RuneError {
		panic("EOF matched")
	}
	rs.writer.WriteRune(r)
	s := string(r)
	rs.end += uint64(len([]byte(s)))
	if rs.isEof {
		return
	}
	if rs.next != 0 {
		rs.next = 0
		return
	}
}
