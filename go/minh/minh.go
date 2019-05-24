package minh

import (
	"fmt"
	"reflect"
	"strings"
	"math"
	"runtime"
	
	minnow "github.com/phil-mansfield/minnow/go"

	"unsafe"
)

const (
	Magic = 0xbaff1ed
	Version = 0
)

const (
	basicFileType  int64 = iota
)

type Writer struct {
	f *minnow.Writer
	blocks int
	cols []Column
	blockSizes []int64
	buf []float32
}

type Column struct {
	Type int64
	Log int32
	Low, High, Dx float32
	Buffer [232]byte
}

type idHeader struct {
	Magic, Version, FileType int64
}

func Create(fname string) *Writer {
	if unsafe.Sizeof(Column{}) != 256 {
		panic(fmt.Sprintf("Sizeof(Column{}) = %d, not 256. Change buffer size.",
			unsafe.Sizeof(Column{})))
	}

	wr := &Writer{
		f: minnow.Create(fname),
	}
	wr.f.Header(idHeader{ Magic, Version, basicFileType })
	return wr
}

func (minh *Writer) Header(names []string, text string, cols []Column) {
	minh.f.Header([]byte(strings.Join(names, "$")))
	minh.f.Header([]byte(text))
	minh.f.Header(cols)
	minh.cols = cols
}

func (minh *Writer) Block(cols []interface{}) {
	if len(cols) != len(minh.cols) {
		panic(fmt.Sprintf("Expected %d columns, got %d.", 
			len(minh.cols), len(cols)))
	}
	for i := range cols {
		colType := minh.cols[i].Type
		if err := minnow.TypeMatch(cols[i], colType); err != nil {
			panic(fmt.Sprintf("Column %d: %s", i, err.Error()))
		}
	}
	
	N := reflect.ValueOf(cols[0]).Len()
	minh.blockSizes = append(minh.blockSizes, int64(N))
	minh.blocks++

	for i := range cols {
		if Ni := reflect.ValueOf(cols[0]).Len(); N != Ni {
			panic(fmt.Sprintf("len(cols[%d]) = %d instead of %d", Ni, i, N))
		}

		colType := minh.cols[i].Type
		switch {
		case colType >= minnow.Int64Group && colType <= minnow.Float32Group:
			minh.f.FixedSizeGroup(colType, N)
			minh.f.Data(cols[i])
		case colType == minnow.IntGroup:
			minh.f.IntGroup(N)
			minh.f.Data(cols[i])
		case colType == minnow.FloatGroup:
			lim := [2]float32{ minh.cols[i].Low, minh.cols[i].High }
			x := cols[i].([]float32)
			minh.buf = expandFloat32(minh.buf, len(x))
			for j := range x {
				minh.buf[j] = x[j]

				if minh.cols[i].Log != 0 {
					minh.buf[j] = float32(math.Log10(float64(minh.buf[j])))
				}
				if minh.buf[j] < minh.cols[i].Low {
					minh.buf[j] = minh.cols[i].Low
				}
				if minh.buf[j] >= minh.cols[i].High {
					minh.buf[j] = math.Nextafter32(
						minh.cols[i].High, float32(math.Inf(-1)),
					)
				}
			}

			minh.f.FloatGroup(N, lim, minh.cols[i].Dx)
			minh.f.Data(minh.buf)
		}
	}
}

func (minh *Writer) Close() {
	minh.f.Header(int64(minh.blocks))
	minh.f.Header(minh.blockSizes)
	minh.f.Close()
}

func expandFloat32(buf []float32, N int) []float32 {
	if cap(buf) >= N { return buf[:N] }
	buf = buf[:cap(buf)]
	return append(buf, make([]float32, N - len(buf))...)
}

func expandInt64(buf []int64, N int) []int64 {
	if cap(buf) >= N { return buf[:N] }
	buf = buf[:cap(buf)]
	return append(buf, make([]int64, N - len(buf))...)
}

type Reader struct {
	Names []string
	Text string
	Columns []Column
	Blocks, Length int
	BlockLengths []int

	f *minnow.Reader
}

func Open(fname string) *Reader {
	f := minnow.Open(fname)
	hd := &idHeader{ }
	f.Header(0, hd)

	if hd.Magic != Magic {
		panic(fmt.Sprintf("%s is not a minh file. Expected magic number " + 
			"%d, but got %d.", fname, Magic, hd.Magic))
	} else if hd.Version < Version {
		panic(fmt.Sprintf("%s written with minh version %d, but reader " + 
			"is version %d.", fname, hd.Version, Version))
	}

	byteNames := make([]byte, f.HeaderSize(1))
	byteText := make([]byte, f.HeaderSize(2))
	cols := make([]Column, f.HeaderSize(3)/int(unsafe.Sizeof(Column{})))
	i64Blocks := int64(0)
	i64BlockLengths := make([]int64, f.HeaderSize(5) / 8)

	f.Header(1, byteNames)
	f.Header(2, byteText)
	f.Header(3, cols)
	f.Header(4, &i64Blocks)
	f.Header(5, i64BlockLengths)
	
	minh := &Reader{
		f: f,
		Names: strings.Split(string(byteNames), "$"),
		Text: string(byteText),
		Columns: cols,
		Blocks: int(i64Blocks),
		BlockLengths: make([]int, len(i64BlockLengths)),
	}

	for i := 0; i < len(i64BlockLengths); i++ {
		minh.BlockLengths[i] = int(i64BlockLengths[i])
		minh.Length += int(i64BlockLengths[i])
	}

	return minh
}

func (rd *Reader) Ints(names []string) map[string][]int64 {
	out := map[string][]int64{ }
	for _, name := range names { out[name] = make([]int64, rd.Length) }

	end := 0

	for b := 0; b < rd.Blocks; b++ {
		start := end
		end = start + rd.BlockLengths[b]
		bOut := map[string][]int64{ }

		for _, name := range names { bOut[name] = out[name][start:end] }
		rd.IntBlock(b, bOut)
	}

	return out
}

func (rd *Reader) Floats(names []string) map[string][]float32 {
	out := map[string][]float32{ }
	for _, name := range names { out[name] = make([]float32, rd.Length) }

	end := 0

	for b := 0; b < rd.Blocks; b++ {
		start := end
		end = start + rd.BlockLengths[b]
		bOut := map[string][]float32{ }

		for _, name := range names { bOut[name] = out[name][start:end] }
		rd.FloatBlock(b, bOut)
	}

	return out
}

func (rd *Reader) IntBlock(b int, out map[string][]int64) {
	runtime.GC()
	for name, arr := range out {
		arr = expandInt64(arr, rd.BlockLengths[b])

		if len(arr) != rd.BlockLengths[b] {
			panic(fmt.Sprintf("Reader.BlockLengths[%d] = %d, but " + 
				"len(out[%s]) = %d", rd.BlockLengths[b], b, name, len(arr)))
		}

		c := findName(name, rd.Names)
		idx := c + b*len(rd.Columns)
		if err := minnow.TypeMatch(arr, rd.Columns[c].Type); err != nil {
			panic(fmt.Sprintf("Column '%s': %s", name, err.Error()))
		}

		rd.f.Data(idx, arr)

		out[name] = arr
	}
}

func (rd *Reader) FloatBlock(b int, out map[string][]float32) {
	runtime.GC()
	for name, arr := range out {
		arr = expandFloat32(arr, rd.BlockLengths[b])

		c := findName(name, rd.Names)
		idx := c + b*len(rd.Columns)
		if err := minnow.TypeMatch(arr, rd.Columns[c].Type); err != nil {
			panic(fmt.Sprintf("Column '%s': %s", name, err.Error()))
		}

		rd.f.Data(idx, arr)

		if rd.Columns[c].Log != 0 {
			for i := range arr {
				arr[i] = float32(math.Pow(10, float64(arr[i])))
			}
		}

		out[name] = arr
	}
}

func (rd *Reader) Close() {
	rd.f.Close()
}

func findName(name string, names []string) int {
	for i := range names {
		if name == names[i] { return i }
	}
	panic(fmt.Sprintf("Name %s not in Reader.Names = %s.", name, names))
}
