package minnow

import (
	"testing"
)

type int64RecordHead struct {
	Magic uint64
	Blocks uint64
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
		f.FixedSizeGroup(Int64Group, len(xs[i]))
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
	lengths := make([]uint64, hd.Blocks)
	f.Header(2, lengths)

	// Read data

	xs = make([][]int64, hd.Blocks)
	for i := range xs {
		xs[i] = make([]int64, lengths[i]) 
		f.Data(i, xs[i])
	}

	return xs, string(bText)
}

type groupRecordHeader struct {
	Blocks, N int64
}

func createGroupRecord(fname string, ix []int32, fx []float64, text string) {
	f := Create(fname)
	defer f.Close()

	in, fn := len(ix) / 4, len(fx) / 2
	intHeader := &groupRecordHeader{ 4, int64(in) }
	floatHeader := &groupRecordHeader{ 2, int64(fn) }
	bText := []byte(text)

	f.Header(intHeader)
	f.FixedSizeGroup(Int32Group, len(ix)/4)
	for i := 0; i < 4; i++ {
		f.Data(ix[i*in: (i+1)*in])
	}

	f.Header(floatHeader)
	f.FixedSizeGroup(Float64Group, len(fx)/2)
	for i := 0; i < 2; i++ {
		f.Data(fx[i*fn: (i+1)*fn])
	}

	f.Header(bText)
}

func readGroupRecord(fname string) ([]int32, []float64, string) {
	f := Open(fname)
	defer f.Close()

	iHd, fHd := &groupRecordHeader{}, &groupRecordHeader{}
	f.Header(0, iHd)
	f.Header(1, fHd)
	bText := make([]byte, f.HeaderSize(2))
	f.Header(2, bText)

	ix, fx := make([]int32, iHd.Blocks*iHd.N), make([]float64, fHd.Blocks*fHd.N)
	for i := int64(0); i < iHd.Blocks; i++ {
		f.Data(int(i), ix[i*iHd.N: (i+1)*iHd.N])
	}
	for i := int64(0); i < fHd.Blocks; i++ {
		f.Data(int(i + iHd.Blocks), fx[i*fHd.N: (i+1)*fHd.N])
	}
	
	return ix, fx, string(bText)
}

func createBitIntRecord(fname string, x1 []int64, x2 [][]int64, x3 []int64) {
	f := Create(fname)
	defer f.Close()

	f.IntGroup(len(x1))
	f.Data(x1)

	f.Header(int64(len(x2)))
	f.IntGroup(len(x2[0]))
	for i := range x2 { f.Data(x2[i]) }

	f.IntGroup(len(x3))
	f.Data(x3)
}

func readBitIntRecord(fname string) ([]int64, [][]int64, []int64) {
	f := Open(fname)
	defer f.Close()

	var x2Len int64
	f.Header(0, &x2Len)

	x1 := make([]int64, f.DataLen(0))
	x2 := make([][]int64, x2Len)
	for i := range x2 { x2[i] = make([]int64, f.DataLen(1 + i))}
	x3 := make([]int64, f.DataLen(int(x2Len) + 1))

	f.Data(0, x1)
	for i := range x2 { f.Data(i+1, x2[i]) }
	f.Data(int(x2Len)+1, x3)

	return x1, x2, x3
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


func TestGroupRecord(t *testing.T) {
	fname := "test_files/group_files.test"
	ix := make([]int32, 20)
	fx := make([]float64, 10)
	for i := range ix { ix[i] = int32(i) }
	for i := range fx { fx[i] = float64(i) / 10 }
	text := "I'm a caaaat"

	createGroupRecord(fname, ix, fx, text)
	rdIx, rdFx, rdText := readGroupRecord(fname)

	if !int32sEq(ix, rdIx) {
		t.Errorf("Written ix = %d, but read ix = %d", ix, rdIx)
	} else if !float64sExactEq(fx, rdFx) {
		t.Errorf("Written fx = %.3g, but read fx = %.3g", fx, rdFx)
	} else if text != rdText {
		t.Errorf("Written text = '%s', but read text = '%s'", text, rdText)
	}
}

func TestBitIntRecord(t *testing.T) {
	fname := "test_files/bit_int_record.test"
	x1 := []int64{100, 101, 102, 104}
	x2 := [][]int64{[]int64{1024, 1024, 1024}, []int64{0, 1023, 500}}
	x3 := []int64{-1000000, -500000}

	createBitIntRecord(fname, x1, x2, x3)
	rdX1, rdX2, rdX3 := readBitIntRecord(fname)
	
	if !int64sEq(x1, rdX1) {
		t.Errorf("Wrote x1 = %d, but read x1 = %d", x1, rdX1)
	} 
	if !int64sEq(x3, rdX3) {
		t.Errorf("Wrote x3 = %d, but read x3 = %d", x3, rdX3)
	} 
	if len(rdX2) != len(x2) {
		t.Errorf("Wrote x2 with length %d, but read x2 with length %d.",
			len(x2), len(rdX2))
	} else {
		for i := range x2 {
			if !int64sEq(x2[i], rdX2[i]) {
				t.Errorf("Wrote xs[%d] = %d, but read x2[%d] = %d.",
					i, x2, i, rdX2)
			}
		}
	}
}

func int32sEq(x, y []int32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func int64sEq(x, y []int64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func float64sExactEq(x, y []float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
