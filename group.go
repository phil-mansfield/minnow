package minnow

import (
	"encoding/binary"
	"os"
)

type group interface {
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
}

func newFixedSizeGroup(startBlock, N, bytes int) *fixedSizeGroup {
	return &fixedSizeGroup{
		*newBlockIndex(startBlock), int64(N), int64(bytes),
	}
}

func newFixedSizeGroupFromTail(f *os.File) *fixedSizeGroup {
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

	return g
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
	return newFixedSizeGroup(startBlock, N, 8)
}
func newInt32Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 4)
}
func newInt16Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 2)
}
func newInt8Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 1)
}
func newUint64Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 8)
}
func newUint32Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 4)
}
func newUint16Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 2)
}
func newUint8Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 1)
}
func newFloat64Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 8)
}
func newFloat32Group(startBlock, N int) group {
	return newFixedSizeGroup(startBlock, N, 4)
}
