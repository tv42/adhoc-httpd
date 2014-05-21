// Package order implements a http.FileSystem that strives to list
// directories in sorted order.
package order

import (
	"container/heap"
	"net/http"
	"os"
)

const WindowSize = 1000

// Minimum number of items to request at a time from wrapped Readdir.
const batchSize = 100

// Order is a http.FileSystem wrapper that strives to list directory
// contents in alphabetical order.
//
// To limit memory consumption, if the directory is larger than
// WindowSize, the Readdir results will only approximate the correct
// order. The listing will contain runs of entries in sorted order,
// where the runs are broken only when entries are seen further than
// WindowSize from their desired ordered location.
type Order struct {
	http.FileSystem
}

var _ = http.FileSystem(Order{})

// Open opens a file. See http.FileSystem method Open.
func (o Order) Open(name string) (http.File, error) {
	f, err := o.FileSystem.Open(name)
	if f != nil {
		f = &file{File: f}
	}
	return f, err
}

type file struct {
	http.File
	window fileHeap
	// if not nil, we're in draining mode, and will return this once
	// window is empty
	err error
}

var _ = http.File(&file{})

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	if f.err != nil && len(f.window) == 0 {
		// drained
		return nil, f.err
	}

	res := []os.FileInfo{}
	for {
		if count > 0 && len(res) == count {
			// reached goal
			return res, nil
		}

		for f.err == nil && len(f.window) < WindowSize {
			// not draining yet, and have room in the window -> fill window
			b := WindowSize - len(f.window)
			if b < batchSize {
				b = batchSize
			}
			fis, err := f.File.Readdir(b)
			for _, fi := range fis {
				heap.Push(&f.window, fi)
			}
			if err != nil {
				// delay error until after draining is done
				f.err = err
				break
			}
		}

		if len(f.window) == 0 {
			// drained the window; delay error until next call
			return res, nil
		}

		// move first (least) item from window to result
		fi := heap.Pop(&f.window).(os.FileInfo)
		res = append(res, fi)
	}
}

type fileHeap []os.FileInfo

var _ = heap.Interface(&fileHeap{})

func (h fileHeap) Len() int           { return len(h) }
func (h fileHeap) Less(i, j int) bool { return h[i].Name() < h[j].Name() }
func (h fileHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *fileHeap) Push(x interface{}) {
	*h = append(*h, x.(os.FileInfo))
}

func (h *fileHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
