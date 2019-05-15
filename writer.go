package minnow

import (
	"encoding/binary"
	"os"
	"reflect"
)

// MinnowWriter represents a new file which minnow blocks can be written into.
type MinnowWriter struct {
	f *os.File

	fileType FileType
	headers, blocks int

    writers []group
    headerOffsets, headerSizes []int64
	groupBlocks []int64
    blockOffsets []int64
}

// minnowHeader is the data block written before any user data is added to the
// files.
type minnowHeader struct {
	magic, version, fileType uint64
	headers, blocks, tailStart uint64
}

func Create(fname string, t FileType) *MinnowWriter {
	f, err := os.Create(fname)
	if err != nil { panic(err.Error()) }

	wr := &MinnowWriter{ f: f, fileType: t }

	// For now we don't need anything in the header: that will be handled in the
	// Close() method.
	err = binary.Write(wr.f, binary.LittleEndian, minnowHeader{})
	if err != nil { panic(err.Error()) }

	return wr
}

func (wr *MinnowWriter) Header(x interface{}) int {
	err := binary.Write(wr.f, binary.LittleEndian, x)
	if err != nil { panic(err.Error()) }

	pos, err := wr.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }
	wr.headerOffsets = append(wr.headerOffsets, pos)
	
	size := int64(reflect.TypeOf(x).Size())
	wr.headerSizes = append(wr.headerSizes, size)

	wr.headers++
	return wr.headers - 1
}

func (wr *MinnowWriter) Int64Group(N int) {
	writer := newInt64Group(wr.blocks, N)
	wr.newGroup(writer)
}

func (wr *MinnowWriter) newGroup(g group) {
	wr.writers = append(wr.writers, g)
	wr.groupBlocks = append(wr.groupBlocks, 0)

	pos, err := wr.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }
	wr.blockOffsets = append(wr.blockOffsets, pos)
}

func (wr *MinnowWriter) Data(x interface{}) int {
	writer := wr.writers[len(wr.writers) - 1]
	writer.writeData(wr.f, x)

	wr.groupBlocks[len(wr.groupBlocks) - 1]++
	wr.blocks++
	return wr.blocks - 1
}

func (wr *MinnowWriter) Close() {
	// Finalize running data.
	
	defer wr.f.Close()

	tailStart, err := wr.f.Seek(0, 1)
	if err != nil { panic(err.Error()) }

	// Write default tail.

	groupSizes := make([]int64, len(wr.writers))
	groupTailSizes := make([]int64, len(wr.writers))
	for i := range groupSizes {
		groupSizes[i] = wr.writers[i].dataBytes()
		groupTailSizes[i] = wr.writers[i].tailBytes()
	}

	tailData := [][]int64{
		wr.headerOffsets, wr.headerSizes,
		wr.blockOffsets, groupSizes, groupTailSizes,
	}

	for _, data := range tailData{
		err = binary.Write(wr.f, binary.LittleEndian, data)
		if err != nil { panic(err.Error())}

	}

	// Write group tail.

	for _, g := range wr.writers {
		g.writeTail(wr.f)
	}

	// Write the header.

	_, err = wr.f.Seek(0, 0)
	if err != nil { panic(err.Error()) }
	
	hd := minnowHeader{
		Magic, Version, uint64(wr.fileType),
		uint64(wr.headers), uint64(wr.blocks), uint64(tailStart),
	}
	binary.Write(wr.f, binary.LittleEndian, hd)
}
