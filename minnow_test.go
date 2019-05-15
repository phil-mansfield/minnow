package minnow

import (
	"fmt"

	"testing"
)

type int64RecordHead struct {
	magic uint64
	blocks uint64
}

func createInt64Record(fname string, xs [][]int64, text string) {
	// Set up your headers and such.

	hd := &int64RecordHead{ 0xdeadbeef, uint64(len(xs)) }
	bText := []byte(text)
	lengths := make([]uint64, len(xs))
	for i := range lengths { lengths[i] = uint64(len(xs[i])) }

	// Create the file

	f := Create(fname)
	defer f.Close()

	f.Header(hd)
	f.Header(bText)
	for i := range xs {
		f.Int64Group()
		f.Data(xs[i])
	}
	f.Header(lengths)
}


func readInt64Record(fname string) (xs [][]int64, text string) {
	// Open and confirm type.

	f := Open(fname)

	// Header stuff.

	hd := &int64RecordHead{ }
	f.Header(0, hd)
	bText := make([]byte, f.HeaderSize(1))
	f.Header(1, bText)
	lengths := make([]uint64, hd.blocks)
	f.Header(2, lengths)

	// Read data

	xs = make([][]int64, hd.blocks)
	for i := range xs {
		xs[i] = make([]int64, lengths[i]) 
		f.Data(i, xs[i])
	}

	return xs, string(bText)
}

func TestInt64Record(t *testing.T) {
	fname := "test_files/int_record.test"
	xs := [][]int64{
		[]int64{1, 2, 3, 4},
		[]int64{5},
		[]int64{6, 7, 8, 9},
		[]int64{10, 11, 12},
	}
	text := "I am a cat and I like to meow."
	
	createInt64Record(fname, xs, text)
	rdXs, rdText := readInt64Record(fname)

	if text != rdText {
		t.Errorf("Written text = '%s', but read text = '%s'", text, rdText)
	}

	if len(xs) != len(rdXs) {
		t.Errorf("Written len(xs) = %d, but read len(xs) = %d.",
			len(xs), len(rdXs))
	}

	for i := range rdXs {
		if !int64sEq(rdXs[i], xs[i]) {
			t.Errorf("Written xs[%d] = %d, but read xs[%d] = %d.",
				i, xs[i], i, rdXs[i])
		}
	}
}


func int64sEq(x, y []int64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
