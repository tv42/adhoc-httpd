package order_test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tv42/adhoc-httpd/order"
)

type fs struct{}

func (fs) Open(name string) (http.File, error) {
	if name == "xyzzy" {
		return &dir{}, nil
	}
	panic("bad test behavior")
}

type dir struct {
	eof bool
}

func (dir) Close() error                                 { return nil }
func (dir) Stat() (os.FileInfo, error)                   { return fileInfo{name: "xyzzy", dir: true}, nil }
func (dir) Read([]byte) (int, error)                     { panic("bad test behavior") }
func (dir) Seek(offset int64, whence int) (int64, error) { panic("bad test behavior") }

func (d *dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.eof {
		return nil, io.EOF
	}
	fis := []os.FileInfo{
		fileInfo{name: "quux"},
		fileInfo{name: "foo"},
		fileInfo{name: "bar"},
	}
	d.eof = true
	return fis, nil
}

type fileInfo struct {
	name string
	dir  bool
}

func (f fileInfo) Name() string       { return f.name }
func (f fileInfo) Size() int64        { return 42 }
func (f fileInfo) Mode() os.FileMode  { return 0755 }
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) IsDir() bool        { return f.dir }
func (f fileInfo) Sys() interface{}   { return nil }

func TestSimple(t *testing.T) {
	o := order.Order{fs{}}
	f, err := o.Open("xyzzy")
	if err != nil {
		t.Fatalf("open xyzzy: %v", err)
	}
	fis, err := f.Readdir(100)
	if err != nil && err != io.EOF {
		t.Fatalf("readdir: %v", err)
	}

	want := []string{
		"bar",
		"foo",
		"quux",
	}
	got := []string{}
	for _, fi := range fis {
		got = append(got, fi.Name())
	}
	if g, e := strings.Join(got, " "), strings.Join(want, " "); g != e {
		t.Errorf("readdir wrong results: %v != %v", g, e)
	}
}
