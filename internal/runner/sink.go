package runner

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"

	"nicograshoff.de/x/optask/internal/fs"
)

type sink struct {
	fs   *fs.FileSystem
	node *fs.Node

	outFile   *os.File
	outSliceW *sliceWriter
	outFileW  *bufio.Writer

	errFile   *os.File
	errFileW  *bufio.Writer
	errSliceW *sliceWriter
}

type sliceWriter struct {
	lines []string
	buf   *bytes.Buffer
}

func newSink(fs *fs.FileSystem) *sink {
	s := new(sink)
	s.fs = fs
	s.node = fs.CreateNodeNow()
	return s
}

func (s *sink) NodeID() string {
	return s.fs.NodeID(s.node)
}

func (s *sink) openStdout() io.Writer {
	file, err := s.fs.Create(s.node, "stdout.txt")
	if err != nil {
		log.Panic(err)
	}

	s.outFile = file
	s.outFileW = bufio.NewWriter(file)
	s.outSliceW = newSliceWriter()

	return io.MultiWriter(s.outFileW, s.outSliceW)
}

func (s *sink) openStderr() io.Writer {
	file, err := s.fs.Create(s.node, "stderr.txt")
	if err != nil {
		log.Panic(err)
	}

	s.errFile = file
	s.errFileW = bufio.NewWriter(file)
	s.errSliceW = newSliceWriter()

	return io.MultiWriter(s.errFileW, s.errSliceW)
}

func (s *sink) close() {
	s.outSliceW.Flush()
	s.outFileW.Flush()
	s.outFile.Close()

	s.errSliceW.Flush()
	s.errFileW.Flush()
	s.errFile.Close()
}

func (s *sink) stdoutLines() []string {
	return s.outSliceW.lines
}

func (s *sink) stderrLines() []string {
	return s.errSliceW.lines
}

func newSliceWriter() *sliceWriter {
	sw := new(sliceWriter)
	sw.lines = make([]string, 0)
	sw.buf = new(bytes.Buffer)
	return sw
}

func (sw *sliceWriter) Write(p []byte) (n int, err error) {
	err = nil
	n = 0
	for _, b := range p {
		sw.buf.WriteByte(b)
		n++
		if b == '\n' {
			sw.Flush()
		}
	}

	return
}

func (sw *sliceWriter) Flush() {
	sw.lines = append(sw.lines, sw.buf.String())
	sw.buf.Reset()
}
