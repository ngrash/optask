package archivesink

import (
	"bufio"
	"io"
	"log"
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/exec"
	"os"
	"sync"
)

type Sink struct {
	fac  *Factory
	node *archive.Node

	outFile   *os.File
	outSliceW *SliceWriter
	outFileW  *bufio.Writer

	errFile   *os.File
	errFileW  *bufio.Writer
	errSliceW *SliceWriter
}

type Factory struct {
	sinks   map[string]*Sink
	sinksMu *sync.Mutex
	fs      *archive.FileSystem
}

func NewFactory(fs *archive.FileSystem) *Factory {
	f := new(Factory)
	f.fs = fs
	f.sinks = make(map[string]*Sink)
	f.sinksMu = new(sync.Mutex)
	return f
}

func (f *Factory) NewSink() exec.Sink {
	f.sinksMu.Lock()
	defer f.sinksMu.Unlock()
	id := archive.NewIdentifierNow()
	sink := new(Sink)
	sink.node = f.fs.CreateNode(id)
	sink.fac = f
	f.sinks[sink.node.String()] = sink
	return sink
}

func (f *Factory) GetOpenSink(sinkID exec.SinkID) *Sink {
	f.sinksMu.Lock()
	defer f.sinksMu.Unlock()
	return f.sinks[sinkID.String()]
}

func (s *Sink) ID() exec.SinkID {
	return s.node
}

func (s *Sink) OpenStdout() io.Writer {
	file, err := s.fac.fs.Create(s.node, "stdout.txt")
	if err != nil {
		log.Panic(err)
	}

	s.outFile = file
	s.outFileW = bufio.NewWriter(file)
	s.outSliceW = NewSliceWriter()

	return io.MultiWriter(s.outFileW, s.outSliceW)
}

func (s *Sink) OpenStderr() io.Writer {
	file, err := s.fac.fs.Create(s.node, "stderr.txt")
	if err != nil {
		log.Panic(err)
	}

	s.errFile = file
	s.errFileW = bufio.NewWriter(file)
	s.errSliceW = NewSliceWriter()

	return io.MultiWriter(s.errFileW, s.errSliceW)
}

func (s *Sink) StdoutLines() []string {
	return s.outSliceW.lines
}

func (s *Sink) StderrLines() []string {
	return s.errSliceW.lines
}

func (s *Sink) Close() {
	s.fac.sinksMu.Lock()
	defer s.fac.sinksMu.Unlock()
	delete(s.fac.sinks, s.node.String())

	s.outSliceW.Flush()
	s.outFileW.Flush()
	s.outFile.Close()

	s.errSliceW.Flush()
	s.errFileW.Flush()
	s.errFile.Close()
}
