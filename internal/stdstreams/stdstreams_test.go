package stdstreams

import (
	"testing"
)

func TestWriters(t *testing.T) {
	l := NewLog()
	l.Stdout().Write([]byte("out\n"))
	l.Stderr().Write([]byte("err\n"))

	lines := l.Lines()
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got: %v", len(lines))
	}

	outLine := lines[0]
	if outLine.Text != "out" {
		t.Errorf("Expected Text == \"out\", got: \"%v\"", outLine.Text)
	}
	if outLine.Stream != Out {
		t.Errorf("Expected Stream == %v, got: %v", Out, outLine.Stream)
	}

	errLine := lines[1]
	if errLine.Text != "err" {
		t.Errorf("Expected Text == \"err\", got: \"%v\"", errLine.Text)
	}
	if errLine.Stream != Err {
		t.Errorf("Expected Stream == %v, got: %v", Err, errLine.Stream)
	}
}

func TestFlush(t *testing.T) {
	l := NewLog()
	l.Stdout().Write([]byte("no newline"))

	lines := l.Lines()
	if len(lines) != 0 {
		t.Errorf("Expected 0 lines, got: %v", len(lines))
	}

	l.Flush()

	lines = l.Lines()
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got: %v", len(lines))
	}
}

func TestMarshal(t *testing.T) {
	l := NewLog()
	l.Stdout().Write([]byte("foo\n"))
	l.Stdout().Write([]byte("bar\n"))

	b, err := l.MarshalBinary()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	l2 := NewLog()
	err = l2.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(l2.Lines()) != 2 {
		t.Errorf("Expected 2 lines, got: %v", len(l2.Lines()))
	}
}
