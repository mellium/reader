package reader

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"testing"
	"time"
)

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
