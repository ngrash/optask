package archivesink

import (
	"bufio"
	"io"
	"log"
	"nicograshoff.de/x/optask/archive"
	"os"
)

type Sink struct {
	fs   *archive.FileSystem
	node *archive.Node

	outFile   *os.File
	outSliceW *SliceWriter
	outFileW  *bufio.Writer

	errFile   *os.File
	errFileW  *bufio.Writer
	errSliceW *SliceWriter
}

func NewSink(fs *archive.FileSystem) *Sink {
	s := new(Sink)
	s.fs = fs
	s.node = fs.CreateNodeNow()
	return s
}

func (s *Sink) NodeID() string {
	return s.fs.NodeID(s.node)
}

func (s *Sink) OpenStdout() io.Writer {
	file, err := s.fs.Create(s.node, "stdout.txt")
	if err != nil {
		log.Panic(err)
	}

	s.outFile = file
	s.outFileW = bufio.NewWriter(file)
	s.outSliceW = NewSliceWriter()

	return io.MultiWriter(s.outFileW, s.outSliceW)
}

func (s *Sink) OpenStderr() io.Writer {
	file, err := s.fs.Create(s.node, "stderr.txt")
	if err != nil {
		log.Panic(err)
	}

	s.errFile = file
	s.errFileW = bufio.NewWriter(file)
	s.errSliceW = NewSliceWriter()

	return io.MultiWriter(s.errFileW, s.errSliceW)
}

func (s *Sink) Close() {
	s.outSliceW.Flush()
	s.outFileW.Flush()
	s.outFile.Close()

	s.errSliceW.Flush()
	s.errFileW.Flush()
	s.errFile.Close()
}

func (s *Sink) StdoutLines() []string {
	return s.outSliceW.lines
}

func (s *Sink) StderrLines() []string {
	return s.errSliceW.lines
}
