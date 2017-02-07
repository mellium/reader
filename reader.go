package reader

import (
	"io"
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
