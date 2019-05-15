package minnow

import (
	"encoding/binary"
	"os"
)

type GroupType int64
const (
	Int64Group GroupType = iota
	Int32Group
	Int16Group
	Int8Group
	Uint64Group
	Uint32Group
	Uint16Group
	Uint8Group
	Float64Group
	Float32Group
)

type group interface {
	groupType() GroupType

	dataBytes() int64
	tailBytes() int64

	writeData(f *os.File, x interface{})
	writeTail(f *os.File)

	blockOffset(b int) int64

	readData(f *os.File, x interface{})
}

var (
	_ group = &fixedSizeGroup{ }
)

////////////////////
// fixedSizeGroup //
////////////////////

type fixedSizeGroup struct {
	blockIndex
	N int64
	typeSize int64
	gt GroupType
}

func newFixedSizeGroup(startBlock, N, bytes int, gt GroupType) *fixedSizeGroup {
	return &fixedSizeGroup{
		*newBlockIndex(startBlock), int64(N), int64(bytes), gt,
	}
}

func newFixedSizeGroupFromTail(f *os.File, gt GroupType) *fixedSizeGroup {
	startBlock := int64(0)
	blocks := int64(0)
	g := &fixedSizeGroup{ }

	err := binary.Read(f, binary.LittleEndian, &g.N)
	if err != nil { panic(err.Error()) }
	err = binary.Read(f, binary.LittleEndian, &startBlock)
	if err != nil { panic(err.Error()) }
	err = binary.Read(f, binary.LittleEndian, &blocks)
	if err != nil { panic(err.Error()) }

	g.blockIndex = *newBlockIndex(int(startBlock))
	for i := int64(0); i < blocks; i++ {
		g.addBlock(g.typeSize*g.N)
	}
	g.gt = gt

	return g
}

func (g *fixedSizeGroup) groupType() GroupType {
	return g.gt
}

func (g *fixedSizeGroup) dataBytes() int64 {
	return 8*g.N
}

func (g *fixedSizeGroup) tailBytes() int64 {
	return 8
}

func (g *fixedSizeGroup) writeData(f *os.File, x interface{}) {
	err := binary.Write(f, binary.LittleEndian, x)
	if err != nil { panic(err.Error()) }
}

func (g *fixedSizeGroup) readData(f *os.File, out interface{}) {
	err := binary.Read(f, binary.LittleEndian, out)
	if err != nil { panic(err.Error()) }
}


func (g *fixedSizeGroup) writeTail(f *os.File) {
	err := binary.Write(f, binary.LittleEndian, g.N)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, g.startBlock)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, g.blocks())
	if err != nil { panic(err.Error()) }
}

/////////////////////////////////
// intances of fixedSizeGroups //
/////////////////////////////////

func newInt64Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 8, Int64Group)
}
func newInt32Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 4, Int32Group)
}
func newInt16Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 2, Int16Group)
}
func newInt8Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 1, Int8Group)
}
func newUint64Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 8, Uint64Group)
}
func newUint32Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 4, Uint32Group)
}
func newUint16Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 2, Uint16Group)
}
func newUint8Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 1, Uint8Group)
}
func newFloat64Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 8, Float64Group)
}
func newFloat32Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 4, Float32Group)
}
