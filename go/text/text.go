package text

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type Reader struct {
	f io.ReadSeeker
	closer io.Closer

	config ReaderConfig

	blocks int
	blockStarts []int64
	blockSizes []int64
	names []string
	buf []byte
}

type ReaderConfig struct {
	Separator byte
	Comment   byte
	MaxBlockSize int64
	MaxItemSize  int64
}

var DefaultReaderConfig = ReaderConfig{
	Separator: byte(' '),
	Comment: byte('#'),
	MaxBlockSize:  5 * (1 << 30),
	MaxItemSize: 100 * (1 << 10),
}

// Open opens a text catalogue and returns a Reader.
func Open(fname string, configOpt ...ReaderConfig) *Reader {
	f, err := os.Open(fname)
	if err != nil { panic(err.Error()) }

	rd := openFromReader(f, configOpt...)

	rd.closer = f
	return rd
}

// openFromReader is the same as Open except it works on a generic ReadSeeker.
// This function exists to make testing easier.
func openFromReader(f io.ReadSeeker, configOpt ...ReaderConfig) *Reader {
	config := DefaultReaderConfig
	if len(configOpt) >= 1 { config = configOpt[0] }

	if config.MaxBlockSize / 2 < config.MaxItemSize {
		panic(fmt.Sprintf("config.MaxBlockSize = %d, but " + 
			"config.MaxLineSize = %d", config.MaxBlockSize,
			config.MaxItemSize))
	}
	
	rd := &Reader{ f: f, config: config }
	rd.findBlocks(readerSize(f))

	return rd
}

// readerSize returns the size of a ReadSeeker
func readerSize(f io.ReadSeeker) int64 {
	pos, err := f.Seek(0, 1)
	if err != nil { panic(err.Error()) }
	size, err := f.Seek(0, 2)
	if err != nil { panic(err.Error())  }
	_, err = f.Seek(pos, 0)
	if err != nil { panic(err.Error()) }

	return size
}

// findBlocks sets blocks, blockStarts, and blockSizes. 
func (rd *Reader) findBlocks(size int64) {
	rd.f.Seek(0, 0)

	rd.blockStarts = []int64{ }
	
	for end := int64(0); end != -1; end = rd.nextBlock(size) {
		rd.blockStarts = append(rd.blockStarts, end)	
	}

	rd.blocks = len(rd.blockStarts)
	rd.blockSizes = make([]int64, rd.blocks)
	for i := 0; i < rd.blocks - 1; i++ {
		rd.blockSizes[i] = rd.blockStarts[i+1] - rd.blockStarts[i]
	}

	rd.blockSizes[rd.blocks - 1] = size - rd.blockStarts[rd.blocks - 1]
}

// nextBlock returns the index of the start of the next block and seeks it.
func (rd *Reader) nextBlock(size int64) int64 {
	curr, err := rd.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }

	// If there are no more blocks in the file
	if curr + int64(rd.config.MaxBlockSize) >= size {
		_, err := rd.f.Seek(0, 2)
		if err != nil { panic(err.Error()) }
		return -1
	}

	// Move to start of search region.
	searchStart := rd.config.MaxBlockSize - rd.config.MaxItemSize + curr
	_, err = rd.f.Seek(searchStart, 0)
	if err != nil { panic(err.Error()) }

	// Read search region.
	rd.buf = expandByte(rd.buf, int(rd.config.MaxItemSize))
	_, err = io.ReadFull(rd.f, rd.buf)
	if err != nil { panic(err.Error()) }

	// Find the next newline.
	delta := int64(bytes.IndexByte(rd.buf, byte('\n')))
	if delta == -1 { panic("config.MaxItemSize too small.") }
	blockEnd := searchStart + delta + 1

	_, err = rd.f.Seek(blockEnd, 0)
	if err != nil { panic(err.Error()) }

	return blockEnd
}

// LinesHeader returns the first lines lines of the file. Must be smaller than
// config.MaxItemSize.
func (rd *Reader) LineHeader(lines int) string { 
	hdLines, _ := rd.headerLines()
	if len(hdLines) <= lines { panic("config.MaxItemSize too small.") }
	return string(bytes.Join(hdLines[:lines], []byte{'\n'}))
}

// CommentHeader returns all the lines that start with comments at the beginning
// of the file. The header must be smaller than config.MaxItemSize.
func (rd *Reader) CommentHeader() string {
	hdLines, nComm := rd.headerLines()
	if len(hdLines) == nComm { panic("config.MaxItemSize too small.") }
	return string(bytes.Join(hdLines[:nComm], []byte{'\n'}))
}

// headerLines returns all the lines in the first config.MaxItemSize bytes of
// the file. The number of comments is also returned.
func (rd *Reader) headerLines() ([][]byte, int) {
	_, err := rd.f.Seek(0, 0)
	if err != nil { panic(err.Error())  }

	bufSize := readerSize(rd.f)
	if bufSize > rd.config.MaxItemSize { bufSize = rd.config.MaxItemSize }

	rd.buf = expandByte(rd.buf, int(bufSize))
	_, err = io.ReadFull(rd.f, rd.buf)
	if err != nil { panic(err.Error()) }

	return split(rd.buf, byte('\n'), rd.config.Comment)
}

// SetNames sets the names associated with each column in the file.
func (rd *Reader) SetNames(names []string) { rd.names = names }

// Blocks returns the number of blocks in the file.
func (rd *Reader) Blocks() int { return rd.blocks }

// Close closes the file.
func (rd *Reader) Close() { rd.closer.Close() }

// Block reads the columns associate with names in block b into the array of
// []int64 and []float32 buffers in out. You'll need to intialize each field in
// out with []int64{} or []float32{} instead of nil. Sorry.
func (rd *Reader) Block(b int, names []string, out []interface{}) {
	if rd.names == nil {
		panic("Must call text.Reader.SetNames() before text.Reader.Blocks()")
	}

	_, err := rd.f.Seek(int64(rd.blockStarts[b]), 0)
	if err != nil { panic(err.Error()) }

	rd.buf = expandByte(rd.buf, int(rd.blockSizes[b]))
	_, err = io.ReadFull(rd.f, rd.buf)
	if err != nil { panic(err.Error()) }

	lines, nComm := split(rd.buf, byte('\n'), rd.config.Comment)
	lines = uncomment(lines, rd.config.Comment, nComm)
	lines = trim(lines, rd.config.Separator)

	for i := range out { out[i] = expandGeneric(out[i], len(lines)) }

	iIdx, iCols, fIdx, fCols := rd.splitByType(names, out)
	parseInt64s(lines, rd.config.Separator, iIdx, iCols)
	parseFloat32s(lines, rd.config.Separator, fIdx, fCols)
}

// splitByType splits the output slices by type (i.e. []int64 and []float32).
func (rd *Reader) splitByType(
	names []string, out []interface{},
) (iIdx []int, iCols [][]int64, fIdx []int, fCols [][]float32) {
	for i := range out {
		idx := rd.nameIndex(names[i])
		switch col := out[i].(type) {
		case []int64:
			iIdx = append(iIdx, idx)
			iCols = append(iCols, col)
		case []float32:
			fIdx = append(fIdx, idx)
			fCols = append(fCols, col)
		default:
			panic(fmt.Sprintf("Type %T can't be used as an output " + 
				"for Block()", col))
		}
	}
	return iIdx, iCols, fIdx, fCols
}

// nameIndex returns the index of the given name.
func (rd *Reader) nameIndex(name string) int {
	for i := range rd.names {
		if strings.ToLower(rd.names[i]) == strings.ToLower(name) {
			return i
		}
	}
	panic(fmt.Sprintf("Name '%s' doesn't match to any columns.", name))
}

// expandByte sets the lengths of x to n. This may involve appending.
func expandByte(x []byte, n int) []byte {
	if cap(x) >= n { return x[:n] }
	x = x[:cap(x)]
	return append(x, make([]byte, n - len(x))...)
}

// expandByte sets the lengths of x to n. This may involve appending.
func expandFloat32(x []float32, n int) []float32 {
	if cap(x) >= n { return x[:n] }
	x = x[:cap(x)]
	return append(x, make([]float32, n - len(x))...)
}

// expandByte sets the lengths of x to n. This may involve appending.
func expandInt64(x []int64, n int) []int64 {
	if cap(x) >= n { return x[:n] }
	x = x[:cap(x)]
	return append(x, make([]int64, n - len(x))...)
}

// expandGeneric sets the lengths of x to n. This may involve appending.
func expandGeneric(x interface{}, n int) interface{} {
	switch col := x.(type) {
	case []int64: return expandInt64(col, n)
	case []float32: return expandFloat32(col, n)
	default:
		panic(fmt.Sprintf("Type %T can't be used as an output " + 
			"for Block()", col))
	}
}
