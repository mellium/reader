package reader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNumRead(t *testing.T) {
	r := strings.NewReader("12345671234567")
	// Read for the tests in two chunks of 7
	mr := io.MultiReader(
		io.LimitReader(r, 7),
		io.LimitReader(r, 7),
	)
	br := bufio.NewReader(mr)

	nrr := numReadReader{
		R: br,
	}

	// One more byte so that if we break the test it will fail anyways.
	for i := 0; i < 2; i++ {
		p := make([]byte, 8)
		n, err := nrr.Read(p)
		switch {
		case err != nil:
			t.Fatalf("Unexpected error while reading: '%v'", err)
		case n != 7:
			t.Fatalf("Read an unexpected ammount, want=7, got=%d", n)
		case p[0] != '1' || p[6] != '7':
			t.Fatalf("Expected to read want=1234567, got=%d", p)
		case nrr.TotalRead != 7*(i+1):
			t.Fatalf("Wrong value for total bytes read; want=%d, got=%d", 7*(i+1), nrr.TotalRead)
		}
	}
}

func TestNumReadByte(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5, 6, 7}
	r := bytes.NewReader(b)
	nrr := numReadReader{
		R: r,
	}
	for i, v := range b {
		bb, err := nrr.ReadByte()
		switch {
		case err != nil:
			t.Fatalf("Unexpected error while reading: '%v'", err)
		case bb != v:
			t.Fatalf("Unexpected byte read; want=%d, got=%d", v, bb)
		case nrr.TotalRead != i+1:
			t.Fatalf("Unexpected value for TotalRead; want=%d, got=%d", i+1, nrr.TotalRead)
		}
	}
}

func TestError(t *testing.T) {
	err := errors.New("Oops")
	er := Error(err)
	if _, e := er.Read(nil); e != err {
		t.Fatalf("Expected original error but got %v", e)
	}
}

func TestBefore(t *testing.T) {
	n := 0
	r := Before(strings.NewReader("Send these, the homeless, tempest-tost to me"), func() error {
		n++
		return nil
	})
	for i := 0; i < 3; i++ {
		p := make([]byte, 4)
		_, err := r.Read(p)
		switch {
		case err != nil:
			t.Fatalf("Read returned unexpected error: %v", err)
		case n > 1:
			t.Fatal("Before func executed more than once!")
		}
	}
}

func TestBeforeError(t *testing.T) {
	in := "I had a lover's quarrel with the world."
	r := strings.NewReader(in)
	oops := errors.New("Oops")
	before := Before(r, func() error {
		return oops
	})
	if _, err := ioutil.ReadAll(before); err != oops {
		t.Fatalf("Unexpected error returned, want=%v, got=%v", oops, err)
	}
	b, err := ioutil.ReadAll(r)
	switch {
	case err != nil:
		t.Fatalf("Unexpected error returned from inner reader: %v", err)
	case string(b) != in:
		t.Fatalf("Failed to read expected value from inner reader, was some of it already read? want='%s', got='%s'", in, string(b))
	}
}

func TestAfter(t *testing.T) {
	oops := errors.New("oops")
	n := 0
	ar := After(Error(io.EOF), func() error {
		n++
		return oops
	})
	for i := 0; i < 3; i++ {
		_, err := ar.Read(nil)
		switch {
		case err != oops:
			t.Fatalf("Unexpected error returned, want=%v, got=%v", oops, err)
		case n != i+1:
			t.Fatalf("Failed to call function after EOF read")
		}
	}
	ar.(*afterReader).r = Error(errors.New("Fail"))
	ar.Read(nil)
	if n > 3 {
		t.Fatalf("Called function even though io.EOF was not returned")
	}
}

type panicConn struct{}

func (panicConn) Read(b []byte) (n int, err error) {
	panic("reader: not implemented")
}

func (panicConn) Write(b []byte) (n int, err error) {
	panic("reader: not implemented")
}

func (panicConn) Close() error {
	panic("reader: not implemented")
}

func (panicConn) LocalAddr() net.Addr {
	panic("reader: not implemented")
}

func (panicConn) RemoteAddr() net.Addr {
	panic("reader: not implemented")
}

func (panicConn) SetDeadline(t time.Time) error {
	panic("reader: not implemented")
}

func (panicConn) SetReadDeadline(t time.Time) error {
	panic("reader: not implemented")
}

func (panicConn) SetWriteDeadline(t time.Time) error {
	panic("reader: not implemented")
}

func TestConn(t *testing.T) {
	c := Conn(panicConn{}, ReaderFunc(func(_ []byte) (int, error) {
		return 10, nil
	}))
	n, err := c.Read(nil)
	switch {
	case n != 10:
		t.Errorf("Got wrong number from conn read, got=%d, want=10", n)
	case err != nil:
		t.Errorf("Did not expect error from Conn's new read method, got: %v", err)
	}
}
