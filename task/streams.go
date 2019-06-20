package task

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type StdStream interface {
	WriteTo(w io.Writer)
	Lines() []string
	Close()
}

// A StdStream implementation that reads from a buffer
type SliceStream struct {
	buf []string
}

func newSliceStream(buf []string) *SliceStream {
	return &SliceStream{buf}
}

func (s *SliceStream) WriteTo(w io.Writer) {
	for _, line := range s.buf {
		fmt.Fprint(w, line)
	}
}

func (s *SliceStream) Lines() []string {
	return s.buf
}

func (s *SliceStream) Close() {}

// A StdStrean implementation that reads from a file
type FileStream struct {
	file *os.File
}

func newFileStream(f *os.File) *FileStream {
	return &FileStream{f}
}

func (s *FileStream) WriteTo(w io.Writer) {
	r := bufio.NewReader(s.file)
	r.WriteTo(w)
}

func (s *FileStream) Lines() []string {
	buf := make([]string, 0)
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		line := scanner.Text()
		buf = append(buf, line)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return buf
}

func (s *FileStream) Close() {
	s.file.Close()
}
