package exec

import "io"

type SinkID interface {
	String() string
}

type stringSinkID struct {
	s string
}

func (id stringSinkID) String() string {
	return id.s
}

func NewSinkID(s string) SinkID {
	return stringSinkID{s}
}

type Sink interface {
	ID() SinkID
	OpenStdout() io.Writer
	OpenStderr() io.Writer
	Close()
}

type SinkFactory interface {
	NewSink() Sink
}
