package runner

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"nicograshoff.de/x/optask/archive"
	"os"
	"sync"
)

type SliceWriter struct {
	lines []string
	buf   *bytes.Buffer
}

type ArchiveSink struct {
	node      *archive.Node
	fac       *ArchiveSinkFactory
	outFile   *os.File
	outSliceW *SliceWriter
	outFileW  *bufio.Writer

	errFile   *os.File
	errFileW  *bufio.Writer
	errSliceW *SliceWriter
}

type ArchiveSinkFactory struct {
	sinks   map[string]*ArchiveSink
	sinksMu *sync.Mutex
	fs      *archive.FileSystem
}

func NewArchiveSinkFactory(fs *archive.FileSystem) *ArchiveSinkFactory {
	f := new(ArchiveSinkFactory)
	f.fs = fs
	f.sinks = make(map[string]*ArchiveSink)
	f.sinksMu = new(sync.Mutex)
	return f
}

func (f *ArchiveSinkFactory) NewSink() Sink {
	f.sinksMu.Lock()
	defer f.sinksMu.Unlock()
	id := archive.NewIdentifierNow()
	sink := new(ArchiveSink)
	sink.node = f.fs.CreateNode(id)
	sink.fac = f
	f.sinks[sink.node.String()] = sink
	return sink
}

func (f *ArchiveSinkFactory) GetOpenSink(sinkID SinkID) *ArchiveSink {
	f.sinksMu.Lock()
	defer f.sinksMu.Unlock()
	return f.sinks[sinkID.String()]
}

func (s *ArchiveSink) ID() SinkID {
	return s.node
}

func (s *ArchiveSink) OpenStdout() io.Writer {
	file, err := s.fac.fs.Create(s.node, "stdout.txt")
	if err != nil {
		log.Panic(err)
	}

	s.outFile = file
	s.outFileW = bufio.NewWriter(file)
	s.outSliceW = NewSliceWriter()

	return io.MultiWriter(s.outFileW, s.outSliceW)
}

func (s *ArchiveSink) OpenStderr() io.Writer {
	file, err := s.fac.fs.Create(s.node, "stderr.txt")
	if err != nil {
		log.Panic(err)
	}

	s.errFile = file
	s.errFileW = bufio.NewWriter(file)
	s.errSliceW = NewSliceWriter()

	return io.MultiWriter(s.errFileW, s.errSliceW)
}

func (s *ArchiveSink) StdoutLines() []string {
	return s.outSliceW.lines
}

func (s *ArchiveSink) StderrLines() []string {
	return s.errSliceW.lines
}

func (s *ArchiveSink) Close() {
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

func NewSliceWriter() *SliceWriter {
	sw := new(SliceWriter)
	sw.lines = make([]string, 0)
	sw.buf = new(bytes.Buffer)
	return sw
}

func (sw *SliceWriter) Write(p []byte) (n int, err error) {
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

func (sw *SliceWriter) Flush() {
	sw.lines = append(sw.lines, sw.buf.String())
	sw.buf.Reset()
}
