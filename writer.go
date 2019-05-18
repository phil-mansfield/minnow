package minnow

import (
	"encoding/binary"
	"os"
)

// MinnowWriter represents a new file which minnow blocks can be written into.
type Writer struct {
	f *os.File

	headers, blocks int

    writers []group
    headerOffsets, headerSizes []int64
	groupBlocks []int64
    groupOffsets []int64
}

// minnowHeader is the data block written before any user data is added to the
// files.
type minnowHeader struct {
	Magic, Version uint64
	Groups, Headers, Blocks uint64
	TailStart int64
}

// Create creates a new minnow file and returns a corresponding Writer.
func Create(fname string) *Writer {
	f, err := os.Create(fname)
	if err != nil { panic(err.Error()) }

	wr := &Writer{ f: f }
	binaryWrite(wr.f, &minnowHeader{})

	return wr
}

// Header writes a header block to the file and returns its header index.
func (wr *Writer) Header(x interface{}) int {
	pos, err := wr.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }
	wr.headerOffsets = append(wr.headerOffsets, pos)
	wr.headerSizes = append(wr.headerSizes, int64(binary.Size(x)))

	binaryWrite(wr.f, x)

	wr.headers++
	return wr.headers - 1
}

// FixedSizeGroup starts a "fixed size" group, meaning that each block only
// contains in16s, uint64, float32s, etc. They are not compressed.
func (wr *Writer) FixedSizeGroup(groupType int64, N int) {
	wr.newGroup(newFixedSizeGroup(wr.blocks, N, groupType))
}

// IntGroup starts an integer group which stores int64s to the minimum
// neccessary precision.
func (wr *Writer) IntGroup(N int) {
	wr.newGroup(newIntGroup(wr.blocks, N))
}

// newGroup starts a new group.
func (wr *Writer) newGroup(g group) {
	wr.writers = append(wr.writers, g)
	wr.groupBlocks = append(wr.groupBlocks, 0)

	pos, err := wr.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }
	wr.groupOffsets = append(wr.groupOffsets, pos)
}

// Data writes a data block to the file within the most recent Group.
func (wr *Writer) Data(x interface{}) int {
	writer := wr.writers[len(wr.writers) - 1]
	writer.writeData(wr.f, x)
	
	wr.groupBlocks[len(wr.groupBlocks) - 1]++
	wr.blocks++
	return wr.blocks - 1
}

// Close writes internal bookkeeping information to the end of the file
// and closes it.
func (wr *Writer) Close() {
	defer wr.f.Close()
	tailStart, err := wr.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }

	// Write default tail.

	groupTypes := make([]int64, len(wr.writers))
	for i := range groupTypes {
		groupTypes[i] = int64(wr.writers[i].groupType())
	}

	tailData := [][]int64{
		wr.headerOffsets, wr.headerSizes, wr.groupOffsets,
		groupTypes, wr.groupBlocks,
	}
	
	for _, data := range tailData {
		binaryWrite(wr.f, data)
	}
	for _, g := range wr.writers {
		g.writeTail(wr.f)
	}

	// Write the header.

	_, err = wr.f.Seek(0, 0)
	if err != nil { panic(err.Error()) }
	
	hd := minnowHeader{
		Magic, Version, uint64(len(wr.writers)),
		uint64(wr.headers), uint64(wr.blocks), tailStart,
	}
	binaryWrite(wr.f, hd)
}

func binaryWrite(f *os.File, data interface{}) {
    err := binary.Write(f, binary.LittleEndian, data)
    if err != nil { panic(err.Error()) }
}
