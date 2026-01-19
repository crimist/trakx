package backup

import (
	"io"
)

// StreamSource wraps an existing reader (e.g., stdin, network stream).
type StreamSource struct {
	Reader      io.Reader
	CloseReader bool // whether to close reader on Close()
	Name        string
}

// NewStreamSource creates a new stream source.
func NewStreamSource(r io.Reader, name string) *StreamSource {
	return &StreamSource{
		Reader:      r,
		CloseReader: false,
		Name:        name,
	}
}

func (s *StreamSource) Open() (io.ReadCloser, error) {
	if s.CloseReader {
		if rc, ok := s.Reader.(io.ReadCloser); ok {
			return rc, nil
		}
	}
	return io.NopCloser(s.Reader), nil
}

func (s *StreamSource) Description() string {
	if s.Name != "" {
		return "stream:" + s.Name
	}
	return "stream:anonymous"
}

// StreamDestination wraps an existing writer (e.g., stdout, network stream).
type StreamDestination struct {
	Writer      io.Writer
	CloseWriter bool // whether to close writer on Close()
	Name        string
}

// NewStreamDestination creates a new stream destination.
func NewStreamDestination(w io.Writer, name string) *StreamDestination {
	return &StreamDestination{
		Writer:      w,
		CloseWriter: false,
		Name:        name,
	}
}

func (d *StreamDestination) Create() (io.WriteCloser, error) {
	if d.CloseWriter {
		if wc, ok := d.Writer.(io.WriteCloser); ok {
			return wc, nil
		}
	}
	return &nopWriteCloser{d.Writer}, nil
}

func (d *StreamDestination) Description() string {
	if d.Name != "" {
		return "stream:" + d.Name
	}
	return "stream:anonymous"
}

type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error {
	return nil
}
