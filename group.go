package minnow

import (
	"os"
)

const (
	Int64Group int64 = iota
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
	groupType() int64

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

func groupFromTail(f *os.File, gt int64) group {
	switch {
	case gt >= Int64Group && gt <= Float64Group:
		return newFixedSizeGroupFromTail(f, gt)
	}
	panic("Unrecognized group type.")
}

////////////////////
// fixedSizeGroup //
////////////////////

type fixedSizeGroup struct {
	blockIndex
	N int64
	typeSize int64
	gt int64
}

func newFixedSizeGroup(startBlock, N, bytes int, gt int64) *fixedSizeGroup {
	return &fixedSizeGroup{
		*newBlockIndex(startBlock), int64(N), int64(bytes), gt,
	}
}

func newFixedSizeGroupFromTail(f *os.File, gt int64) *fixedSizeGroup {
	startBlock := int64(0)
	blocks := int64(0)
	g := &fixedSizeGroup{ }

	binaryRead(f, &g.N)
	binaryRead(f, &startBlock)
	binaryRead(f, &blocks)

	g.blockIndex = *newBlockIndex(int(startBlock))
	for i := int64(0); i < blocks; i++ {
		g.addBlock(g.typeSize*g.N)
	}
	g.gt = gt

	return g
}

func (g *fixedSizeGroup) groupType() int64 {
	return g.gt
}

func (g *fixedSizeGroup) dataBytes() int64 {
	return 8*g.N
}

func (g *fixedSizeGroup) tailBytes() int64 {
	return 8
}

func (g *fixedSizeGroup) writeData(f *os.File, x interface{}) {
	binaryWrite(f, x)
	g.addBlock(g.typeSize*g.N)
}

func (g *fixedSizeGroup) readData(f *os.File, out interface{}) {
	binaryRead(f, out)
}


func (g *fixedSizeGroup) writeTail(f *os.File) {
	binaryWrite(f, g.N)
	binaryWrite(f, g.startBlock)
	binaryWrite(f, g.blocks())
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
