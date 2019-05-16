package minnow

import (
	"encoding/binary"
	"fmt"
	"os"
)

//////////////////
// MinnowReader //
//////////////////

// MinnowReader represents an open minnow file.
type MinnowReader struct {
	f *os.File

	groups, headers, blocks int

	readers []group
	blockIndex []int
	
	headerOffsets, headerSizes []int64
	groupOffsets, groupSizes, groupHeaderSizes []int64
	groupTypes []int64
}

// Open opens a minnow file.
func Open(fname string) *MinnowReader {
	f, err := os.Open(fname)

	// Read header

	minHd := &minnowHeader{}
	binaryRead(f, minHd)

	// Check that this is a file we can actually read.

	if minHd.Magic != Magic {
		panic(fmt.Sprintf("%s is not a minnow file. Magic number is %x, " + 
			"not %x.", fname, minHd.Magic, Magic))
	} else if minHd.Version != Version {
		panic(fmt.Sprintf("%s was written with minnow verison %d, but this " +
			"code has version %d. See the github page for instrucitons on " + 
			"retrieving a specific version.", fname, minHd.Version, Version))
	}

	rd := &MinnowReader{
		f: f, groups: int(minHd.Groups),
		headers: int(minHd.Headers), blocks: int(minHd.Blocks),
	}

	// Read tail data

	_, err = f.Seek(minHd.TailStart, 0)
	if err != nil { panic(err.Error()) }	

	rd.headerOffsets = make([]int64, rd.headers)
	rd.headerSizes = make([]int64, rd.headers)
	rd.groupOffsets = make([]int64, rd.groups)
	rd.groupTypes = make([]int64, rd.groups)
	groupBlocks := make([]int64, rd.groups)

	tailData := [][]int64{
		rd.headerOffsets, rd.headerSizes, rd.groupOffsets,
		rd.groupTypes, groupBlocks,
	}

	// Read group data

	for _, data := range tailData {
		binaryRead(f, data)
	}
	for i := 0; i < rd.groups; i++ {
		rd.readers = append(rd.readers, groupFromTail(f, rd.groupTypes[i]))
	}

	rd.blockIndex = make([]int, rd.blocks)
	i := 0
	for j := range groupBlocks {
		for k := 0; k < int(groupBlocks[j]); k++ {
			rd.blockIndex[i] = j
			i++
		}
	}

	return rd
}


// Header reads the ith header in the minnow file.
func (rd *MinnowReader) Header(i int, out interface{}) {
	if binary.Size(out) != int(rd.headerSizes[i]) {
		panic(fmt.Sprintf("Header buffer has size %d, but written header " + 
			"has size %d.", binary.Size(out), rd.headerSizes[i]))
	}

	_, err := rd.f.Seek(rd.headerOffsets[i], 0)
	if err != nil { panic(err.Error()) }
	binaryRead(rd.f, out)
}

// HeaderSize returns the number of bytes in ith header in the file.
func (rd *MinnowReader) HeaderSize(i int) int {
	return int(rd.headerSizes[i])
}

// Blocks returns the number of data blocks in the file.
func (rd *MinnowReader) Blocks() int {
	return rd.blocks
}

// Data reads the bth data block in the file.
func (rd *MinnowReader) Data(b int, out interface{}) {
	i := rd.blockIndex[b]
	_, err := rd.f.Seek(rd.groupOffsets[i], 0)
	if err != nil { panic(err.Error()) }
	_, err = rd.f.Seek(rd.readers[i].blockOffset(b), 1)
	if err != nil { panic(err.Error()) }

	rd.readers[i].readData(rd.f, out)
}

// DataType returns an integer representing the 
func (rd *MinnowReader) DataType(b int) int64 {
	return rd.groupTypes[rd.blockIndex[b]]
}


// Close closes the file.
func (rd *MinnowReader) Close() {
	rd.f.Close()
}

func binaryRead(f *os.File, data interface{}) {
	err := binary.Read(f, binary.LittleEndian, data)
	if err != nil { panic(err.Error()) }
}
