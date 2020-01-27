package stdstreams

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"sync"
	"time"
)

const (
	Out = 1
	Err = 2
)

type Line struct {
	Stream int
	Time   time.Time
	Text   string
}

type Log struct {
	lines []Line
	mutex sync.Mutex
	outW  *bufferedLineWriter
	errW  *bufferedLineWriter
}

func NewLog() *Log {
	l := &Log{
		make([]Line, 0),
		sync.Mutex{},
		nil,
		nil,
	}

	l.outW = l.makeWriter(Out)
	l.errW = l.makeWriter(Err)

	return l
}

func (l *Log) Stdout() io.Writer {
	return l.outW
}

func (l *Log) Stderr() io.Writer {
	return l.errW
}

func (l *Log) makeWriter(stream int) *bufferedLineWriter {
	return newBufferedLineWriter(func(text string) {
		l.writeLine(stream, text)
	})
}

func (l *Log) writeLine(stream int, text string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.lines = append(l.lines, Line{stream, time.Now(), text})
}

func (l *Log) Lines() []Line {
	return l.lines
}

func (l *Log) Flush() {
	l.errW.Flush()
	l.outW.Flush()
}

func (l *Log) JSON(skip int) ([]byte, error) {
	return json.Marshal(l.lines[skip:])
}

func (l *Log) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(l.lines); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (l *Log) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	return dec.Decode(&l.lines)
}

type bufferedLineWriter struct {
	buf *bytes.Buffer
	fn  func(text string)
}

func newBufferedLineWriter(fn func(text string)) *bufferedLineWriter {
	return &bufferedLineWriter{new(bytes.Buffer), fn}
}

func (w *bufferedLineWriter) Write(p []byte) (int, error) {
	n := 0
	for _, b := range p {
		n++
		if b == '\n' {
			w.fn(w.buf.String())
			w.buf.Reset()
		} else {
			w.buf.WriteByte(b)
		}
	}
	return n, nil
}

func (w *bufferedLineWriter) Flush() {
	if w.buf.Len() > 0 {
		w.fn(w.buf.String())
		w.buf.Reset()
	}
}
