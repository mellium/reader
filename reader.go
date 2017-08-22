package reader // import "mellium.im/reader"

import (
	"io"
	"net"
	"sync"
)

// Error returns a reader that always returns the given error for all calls to
// Read.
func Error(err error) io.Reader {
	return &errReader{
		err: err,
	}
}

type errReader struct {
	err error
}

func (er *errReader) Read(p []byte) (int, error) {
	return 0, er.err
}

// Before returns an io.Reader that proxies calls to Read and executes the given
// function exactly once before the first call.
// If the function errors, the error is returned and the call to Read is never
// proxied to the inner io.Reader (subsequent calls to Read will still be
// proxied).
// Because no call to Read returns until the one call to f returns, if f causes
// Read to be called, it will deadlock.
// If f panics, Read considers it to have returned; future calls of Read return
// without calling f.
// For more information see the documentation for sync.Once.
func Before(r io.Reader, f func() error) io.Reader {
	return &beforeReader{
		r:    r,
		f:    f,
		once: &sync.Once{},
	}
}

type beforeReader struct {
	r    io.Reader
	f    func() error
	once *sync.Once
}

func (br *beforeReader) Read(p []byte) (n int, err error) {
	br.once.Do(func() {
		err = br.f()
	})
	if err != nil {
		return
	}
	return br.r.Read(p)
}

func (br *beforeReader) Reset(r io.Reader) {
	br.r = r
	br.once = &sync.Once{}
}

// After returns an io.Reader that proxies to another Reader and calls f after
// each Read that returns an io.EOF error.
// If io.EOF is never returned, f is never called and if io.EOF is returned
// multiple times, f will be called multiple times.
// If f returns an error, it is returned from the call to Read instead of
// io.EOF.
func After(r io.Reader, f func() error) io.Reader {
	return &afterReader{
		r: r,
		f: f,
	}
}

type afterReader struct {
	r io.Reader
	f func() error
}

func (ar *afterReader) Read(p []byte) (n int, err error) {
	n, err = ar.r.Read(p)
	if err == io.EOF {
		if e := ar.f(); e != nil {
			return n, e
		}
	}
	return n, err
}

// TODO: I wrote this, then ended up not using it. Expose it anyways, or remove?

// numReadReader keeps track of the number of bytes that have been read during
// the lifetime of the reader.
type numReadReader struct {
	R interface {
		io.ByteReader
		io.Reader
	}
	TotalRead int
}

func (nrr *numReadReader) Read(p []byte) (n int, err error) {
	n, err = nrr.R.Read(p)
	nrr.TotalRead += n
	return
}

func (nrr *numReadReader) ReadByte() (byte, error) {
	b, err := nrr.R.ReadByte()
	nrr.TotalRead++
	return b, err
}

// Conn replaces the Read method of c with r.Read.
// Generally, r wraps the Read method of c.
func Conn(c net.Conn, r io.Reader) net.Conn {
	return &conn{
		Conn: c,
		r:    r,
	}
}

type conn struct {
	net.Conn
	r io.Reader
}

func (c *conn) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

// ReaderFunc is an adapter to allow the use of ordinary functions as
// io.Readers.  If f is a function with the appropriate signature, ReaderFunc(f)
// is an io.Reader that calls f.
type ReaderFunc func(p []byte) (n int, err error)

// Read calls f.
func (f ReaderFunc) Read(p []byte) (n int, err error) {
	return f(p)
}
