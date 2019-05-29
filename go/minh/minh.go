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
	basicFileType    int64 = iota
	boundaryFileType
)

const (
	Int64 int64 = iota
	Int32
	Int16
	Int8
	Uint64
	Uint32
	Uint16
	Uint8
	Float64
	Float32
	Int
	Float
)

type Writer struct {
	f *minnow.Writer
	blocks int
	cols []Column
	blockSizes []int64
	buf []float32
	l, boundary float32
	cells int
}

type Column struct {
	Type int64
	Log int32
	Low, High, Dx float32
	Buffer [232]byte
}

func (c Column) String() string {
	return fmt.Sprintf("{Type: %s, Log: %t, Range: (%g %g), Dx: %g}",
		minnow.GroupNames[c.Type], c.Log != 0, c.Low, c.High, c.Dx)
}
func (c Column) GoString() string { return c.String() }

type idHeader struct {
	Magic, Version, FileType int64
}

type geometry struct {
	L, Boundary float32
	Cells int64
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
	wr.f.Header(Boundary{ })
	return wr
}

func (minh *Writer) Header(names []string, text string, cols []Column) {
	minh.f.Header([]byte(strings.Join(names, "$")))
	minh.f.Header([]byte(text))
	minh.f.Header(cols)
	minh.cols = cols
}

func (minh *Writer) Geometry(L, boundary float32, cells int) {
	minh.l, minh.boundary, minh.cells = L, boundary, cells 
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
	minh.f.Header(boundary{ minh.l, minh.boundary, int64(minh.cells) })
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
	L, Boundary float32
	Cells int


	f *minnow.Reader
	fileType int64
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
	geom := &geometry{ }
	i64Blocks := int64(0)
	i64BlockLengths := make([]int64, f.HeaderSize(6) / 8)
	
	f.Header(1, byteNames)
	f.Header(2, byteText)
	f.Header(3, cols)
	f.Header(4, geom)
	f.Header(5, &i64Blocks)
	f.Header(6, i64BlockLengths)
	
	minh := &Reader{
		f: f,
		fileType: hd.FileType,
		Names: strings.Split(string(byteNames), "$"),
		Text: string(byteText),
		Columns: cols,
		Blocks: int(i64Blocks),
		BlockLengths: make([]int, len(i64BlockLengths)),
		L: geom.L,
		Boundary: geom.Boundary,
		Cells: geom.Cells,
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

		var idx int
		c := findName(name, rd.Names)
		if rd.fileType == basicFileType {
			idx = c + b*len(rd.Columns)
		} else {
			idx = c*rd.Blocks + b
		}

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
		var idx int
		if rd.fileType == basicFileType {
			idx = c + b*len(rd.Columns)
		} else {
			idx = c*rd.Blocks + b
		}

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
