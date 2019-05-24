package minh

import (
	"fmt"
	"reflect"
	"strings"
	"math"
	"runtime"
	
	"github.com/phil-mansfield/minnow"

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
	ColumnType int64
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
		colType := minh.cols[i].ColumnType
		if err := minnow.TypeMatch(cols[i], colType); err != nil {
			panic(fmt.Sprintf("Column %d: %s", err.Error()))
		}
	}
	
	N := reflect.ValueOf(cols[0]).Len()
	minh.blockSizes = append(minh.blockSizes, int64(N))

	for i := range cols {
		if Ni := reflect.ValueOf(cols[0]).Len(); N != Ni {
			panic(fmt.Sprintf("len(cols[%d]) = %d instead of %d", Ni, i, N))
		}

		colType := minh.cols[i].ColumnType
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
			for i := range x {
				minh.buf[i] = float32(math.Log10(float64(x[i])))
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
	if cap(buf) <= N { return buf[:N] }
	buf = buf[:cap(buf)]
	return append(buf, make([]float32, N - len(buf))...)
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
	i64BlockSizes := make([]int64, f.HeaderSize(5) / 8)
	
	f.Header(1, byteNames)
	f.Header(2, byteText)
	f.Header(3, cols)
	f.Header(4, &i64Blocks)
	f.Header(5, i64BlockSizes)
	
	minh := &Reader{
		Names: strings.Split(string(byteNames), "$"),
		Text: string(byteText),
		Columns: cols,
		Blocks: int(i64Blocks),
		BlockLengths: make([]int, len(cols)),
	}

	for i := 0; i < len(cols); i++ {
		minh.BlockLengths[i] = int(i64BlockSizes[i])
		minh.Length += int(i64BlockSizes[i])
	}

	return minh
}

func (rd *Reader) ReadInt(names []string) map[string][]int64 {
	out := map[string][]int64{ }
	for _, name := range names { out[name] = make([]int64, rd.Length) }

	end := 0

	for b := 0; b < rd.Blocks; b++ {
		start := end
		end = start + rd.BlockLengths[b]
		bOut := map[string][]int64{ }

		for _, name := range names { bOut[name] = out[name][start:end]}
		rd.ReadIntBlock(b, bOut)
	}

	return out
}

func (rd *Reader) ReadFloat([]string) map[string][]float32 {
	out := map[string][]float32{ }
	for _, name := range rd.Names { out[name] = make([]float32, rd.Length) }

	end := 0

	for b := 0; b < rd.Blocks; b++ {
		start := end
		end = start + rd.BlockLengths[b]
		bOut := map[string][]float32{ }

		for _, name := range rd.Names { bOut[name] = out[name][start:end] }
		rd.ReadFloatBlock(b, bOut)
	}

	return out
}

func (rd *Reader) ReadIntBlock(b int, out map[string][]int64) {
	runtime.GC()
	for name, arr := range out {
		if len(arr) != rd.BlockLengths[b] {
			panic(fmt.Sprintf("Reader.BlockLengths[%d] = %d, but " + 
				"len(out[%s]) = %d", rd.BlockLengths[b], b, name, len(arr)))
		}

		idx := findName(name, rd.Names) + b*len(rd.Columns)
		rd.f.Data(idx, out)
	}
}

func (rd *Reader) ReadFloatBlock(b int, out map[string][]float32) {
	runtime.GC()
	for name, arr := range out {
		if len(arr) != rd.BlockLengths[b] {
			panic(fmt.Sprintf("Reader.BlockLengths[%d] = %d, but " + 
				"len(out[%s]) = %d", rd.BlockLengths[b], b, name, len(arr)))
		}

		idx := findName(name, rd.Names) + b*len(rd.Columns)
		rd.f.Data(idx, out)
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
