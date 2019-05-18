package minnow

import (
	"os"
	"github.com/phil-mansfield/minnow/bit"
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
	IntGroup
)

var fixedSizeBytes = []int{
	8, 4, 2, 1, 8, 4, 2, 1, 8, 4, 
}

type group interface {
	groupType() int64
	length(b int) int

	writeData(f *os.File, x interface{})
	writeTail(f *os.File)

	blockOffset(b int) int64

	readData(f *os.File, b int, x interface{})
}

var (
	_ group = &fixedSizeGroup{ }
)

func groupFromTail(f *os.File, gt int64) group {
	switch {
	case gt >= Int64Group && gt <= Float64Group:
		return newFixedSizeGroupFromTail(f, gt)
	case gt == IntGroup:
		return newIntGroupFromTail(f)
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

func newFixedSizeGroup(startBlock, N int, gt int64) *fixedSizeGroup {
	return &fixedSizeGroup{
		*newBlockIndex(startBlock), int64(N),
		int64(fixedSizeBytes[gt]), gt,
	}
}

func newFixedSizeGroupFromTail(f *os.File, gt int64) *fixedSizeGroup {
	startBlock := int64(0)
	blocks := int64(0)
	g := &fixedSizeGroup{ typeSize: int64(fixedSizeBytes[gt]) }

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

func (g *fixedSizeGroup) length(b int) int {
	return int(g.N)
}

func (g *fixedSizeGroup) writeData(f *os.File, x interface{}) {
	binaryWrite(f, x)
	g.addBlock(g.typeSize*g.N)
}

func (g *fixedSizeGroup) readData(f *os.File, b int, out interface{}) {
	binaryRead(f, out)
}


func (g *fixedSizeGroup) writeTail(f *os.File) {
	binaryWrite(f, g.N)
	binaryWrite(f, g.startBlock)
	binaryWrite(f, g.blocks())
}

///////////////
// IntGroup //
///////////////

// intGroup is a group which dynamically determines the range of the input
// integers and stores using the minimum possible precision. This group is used
// as a component of several other, more comlicated groups.
type intGroup struct {
	blockIndex
	N int64
	ab *bit.ArrayBuffer
	mins, bits []int64
}

func newIntGroup(startBlock, N int) group {
	return &intGroup{
		blockIndex: *newBlockIndex(startBlock), N: int64(N),
		ab: &bit.ArrayBuffer{ },
	}
}

func newIntGroupFromTail(f *os.File) group {
	g := &intGroup{ }
	var startBlock, blocks, min, bits int64
	g.ab = &bit.ArrayBuffer{ }

	read := func() (x []int64) {
		binaryRead(f, &min)
		binaryRead(f, &bits)

		buf := g.ab.Read(f, int(bits), int(blocks))
		out := make([]int64, blocks)
		for i := range out { out[i] = min + int64(buf[i] )}
		return out
	}

	binaryRead(f, &g.N)
	binaryRead(f, &startBlock)
	binaryRead(f, &blocks)
	g.mins = read()
	g.bits = read()

	g.blockIndex = *newBlockIndex(int(startBlock))
	for i := range g.bits {
		g.addBlock(int64(bit.ArrayBytes(int(g.bits[i]), int(g.N))))
	}

	return g
}

func (g *intGroup) writeTail(f *os.File) {
	write := func(x []int64) {
		buf := g.ab.Uint64(len(x))
		min := int64Min(x)
		for i := range x { buf[i] = uint64(x[i] - min) }
		bits := g.ab.Bits(buf)

		binaryWrite(f, min)
		binaryWrite(f, int64(bits))
		g.ab.Write(f, buf, bits)
	}

	binaryWrite(f, int64(g.N))
	binaryWrite(f, g.startBlock)
	binaryWrite(f, g.blocks())
	write(g.mins)
	write(g.bits)
}

func (g *intGroup) groupType() int64 {
	return IntGroup
}

func (g *intGroup) length(b int) int {
	return int(g.N)
}

func (g *intGroup) writeData(f *os.File, x interface{}) {
	data := x.([]int64)
	min := int64Min(data)
	
	buf := g.ab.Uint64(len(data))
	for i := range buf { buf[i] = uint64(data[i] - min) }
	bits := g.ab.Bits(buf)
	g.ab.Write(f, buf, bits)
	
	g.mins = append(g.mins, min)
	g.bits = append(g.bits, int64(bits))

	g.addBlock(int64(bit.ArrayBytes(bits, int(g.N))))
}

func (g *intGroup) readData(f *os.File, b int, x interface{}) {	
	out := x.([]int64)
	bIdx := b - int(g.startBlock)
	bits, min := g.bits[bIdx], g.mins[bIdx]
	buf := g.ab.Read(f, int(bits), int(g.N))
	for i := range buf { out[i] = min + int64(buf[i]) }
}

func int64Min(x []int64) int64 {
	min := x[0]
	for i := range x {
		if x[i] < min { min = x[i] }
	}
	return min
}

func uint64Min(x []uint64) uint64 {
	min := x[0]
	for i := range x {
		if x[i] < min { min = x[i] }
	}
	return min
}
