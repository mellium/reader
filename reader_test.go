// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package reader_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"mellium.im/reader"
)

func TestError(t *testing.T) {
	err := errors.New("Oops")
	er := reader.Error(err)
	if _, e := er.Read(nil); e != err {
		t.Fatalf("Expected original error but got %v", e)
	}
}

func TestBefore(t *testing.T) {
	n := 0
	r := reader.Before(strings.NewReader("Send these, the homeless, tempest-tost to me"), func() error {
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
	before := reader.Before(r, func() error {
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

var oops = errors.New("oops")

type afterTestData struct {
	p   []byte
	n   int
	err error
}

var afterTestCases = [...]struct {
	r    io.Reader
	f    func(int, error) (int, error)
	data []afterTestData
}{
	0: {
		r: reader.Error(io.EOF),
		data: []afterTestData{
			{p: nil, n: 0, err: io.EOF},
			{p: nil, n: 0, err: io.EOF},
		},
	},
	1: {
		r: reader.Error(io.EOF),
		f: func(n int, err error) (int, error) {
			return 2, oops
		},
		data: []afterTestData{
			{p: nil, n: 2, err: oops},
			{p: nil, n: 2, err: oops},
		},
	},
	2: {
		r: strings.NewReader("test"),
		data: []afterTestData{
			{p: make([]byte, 4), n: 4, err: nil},
			{p: nil, n: 0, err: io.EOF},
		},
	},
}

func TestAfter(t *testing.T) {
	for i, tc := range afterTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ar := reader.After(tc.r, tc.f)
			for c, d := range tc.data {
				n, err := ar.Read(d.p)
				if n != d.n {
					t.Errorf("Wrong number of bytes on read %d: want=%d, got=%d", c, d.n, n)
				}
				if err != d.err {
					t.Errorf("Wrong error on read %d: want=%q, got=%q", c, d.err, err)
				}
			}
		})
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
	c := reader.Conn(panicConn{}, reader.Func(func(_ []byte) (int, error) {
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
