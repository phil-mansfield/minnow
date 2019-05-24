package minnow

import (
	"fmt"
	"math"
	"math/rand"
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
	FloatGroup
)

var (
	groupNames = []string{
		"Int64Group",
		"Int32Group",
		"Int16Group",
		"Int8Group",
		"Uint64Group",
		"Uint32Group",
		"Uint16Group",
		"Uint8Group",
		"Float64Group",
		"Float32Group",
		"IntGroup",
		"FloatGroup",
	}
)

func TypeMatch(x interface{}, gt int64) error {
	f := func(s string) error {
		return fmt.Errorf("Got type %s for group %s.", s, groupNames[gt])
	}
	switch v := x.(type) {
	case []int64:
		_ = v // To get type switching to work
		if !(gt == Int64Group || gt == IntGroup) { return f("[]int64") }
	case []int32:
		if !(gt == Int32Group) { return f("[]int32") }
	case []int16:
		if !(gt == Int16Group) { return f("[]int16") }
	case []int8:
		if !(gt == Int8Group) { return f("[]int8") }
	case []uint64:
		if !(gt == Uint64Group) { return f("[]uint64") }
	case []uint32:
		if !(gt == Uint32Group) { return f("[]uint32") }
	case []uint16:
		if !(gt == Uint16Group) { return f("[]int16") }
	case []uint8:
		if !(gt == Uint8Group) { return f("[]int8") }
	case []float64:
		if !(gt == Float64Group) { return f("[]float64") }
	case []float32:
		if !(gt == Float32Group || gt == FloatGroup) { return f("[]float32") }
	}
	return nil
}

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
	case gt >= Int64Group && gt <= Float32Group:
		return newFixedSizeGroupFromTail(f, gt)
	case gt == IntGroup:
		return newIntGroupFromTail(f)
	case gt == FloatGroup:
		return newFloatGroupFromTail(f)
	}
	panic(fmt.Sprintf("Unrecognized group type, %d.", gt))
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

/////////////////
// FloatGroup //
/////////////////

type floatGroup struct {
	ig *intGroup
	low, high float32
	pixels int64
	periodic uint8
	buf []int64
}

func newFloatGroup(
	startBlock, N int, low, high float32, pixels int64, periodic bool,
) group {
	u8Periodic := uint8(0)
	if periodic { u8Periodic = 1 }
	return &floatGroup{
		ig: newIntGroup(startBlock, N).(*intGroup),
		low: low, high: high, pixels: pixels, periodic: u8Periodic,
	}
}

func (g *floatGroup) groupType() int64 {
	return FloatGroup
}
func (g *floatGroup) length(b int) int {
	return g.ig.length(b)
}

func (g *floatGroup) blockOffset(b int) int64 {
	return g.ig.blockOffset(b)
}

func (g *floatGroup) readData(f *os.File, b int, x interface{}) {
	out := x.([]float32)
	g.buf = resizeInt64(g.buf, int(g.ig.N))
	g.ig.readData(f, b, g.buf)
	if g.periodic == 1 { bound(g.buf, 0, g.pixels) }

	L := g.high - g.low
	dx := L / float32(g.pixels)
	for i := range g.buf {
		out[i] = dx*float32(float64(g.buf[i]) + rand.Float64()) + g.low
	}
}

func (g *floatGroup) writeData(f *os.File, x interface{}) {
	data := x.([]float32)
	g.buf = resizeInt64(g.buf, int(g.ig.N))

	dx := (g.high - g.low) / float32(g.pixels)
	
	for i := range g.buf {
		g.buf[i] = int64(math.Floor(float64((data[i] - g.low) / dx)))
	}
	if g.periodic == 1 {
		min := periodicMin(g.buf, g.pixels)
		bound(g.buf, min, g.pixels)
	}

	g.ig.writeData(f, g.buf)
}
func (g *floatGroup) writeTail(f *os.File) {
	g.ig.writeTail(f)
	binaryWrite(f, g.low)
	binaryWrite(f, g.high)
	binaryWrite(f, g.pixels)
	binaryWrite(f, g.periodic)
}

func newFloatGroupFromTail(f *os.File) group {
	g := &floatGroup{ }
	g.ig = newIntGroupFromTail(f).(*intGroup)
	binaryRead(f, &g.low)
	binaryRead(f, &g.high)
	binaryRead(f, &g.pixels)
	binaryRead(f, &g.periodic)
	return g
}

///////////////////////
// utility functions //
///////////////////////

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

func resizeInt64(x []int64, n int) []int64 {
	if cap(x) >= n { return x[:n] }
	x = x[:cap(x)]
	return append(x, make([]int64, n - len(x))...)
}

func bound(x []int64, min, pixels int64) {
	for i := range x {
		if x[i] < min {
			x[i] += pixels
		} else if x[i] >=  min + pixels {
			x[i] -= pixels
		}
	}
}

func periodicMin(x []int64, pixels int64) int64 {	
	x0, width := x[0], int64(1)
		
	for _, xi := range x {
		x1 := x0 + width - 1
		if x1 >= pixels { x1 -= pixels }

		d0 := periodicDistance(xi, x0, pixels)
		d1 := periodicDistance(xi, x1, pixels)

		if d0 > 0 && d1 < 0 { continue }

		if d1 > -d0 {
			width += d1
		} else {
			x0 += d0
			if x0 < 0 { x0 += pixels }
			width -= d0
		}

		if width > pixels/2 { return 0 }
	}

	return x0
}

// periodicDistance computes the distance from x0 to x.
func periodicDistance(x, x0, pixels int64) int64 {
	d := x - x0
	if d >= 0 {
		if d > pixels - d { return d - pixels }
	} else {
		if d < -(d + pixels) { return pixels + d }
	}
	return d
}
