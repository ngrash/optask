package archivesink

import "bytes"

type SliceWriter struct {
	lines []string
	buf   *bytes.Buffer
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
