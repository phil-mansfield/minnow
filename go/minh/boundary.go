package minh

import (
	"fmt"
	"strings"
	
	minnow "github.com/phil-mansfield/minnow/go"
)

type BoundaryWriter struct {
	Writer

	names []string
	cols []Column

	colLength int
	scaledBoundary float32

	cellIndex [][]int
	i64Buf []int64
	f32Buf []float32

	// vector buffers
	vec [3]float32
	idx, sum [3]int
}

func CreateBoundary(fname string) *BoundaryWriter {
	wr := &BoundaryWriter{ }
	wr.create(fname, boundaryFileType)
	return wr
}

func (minh *BoundaryWriter) Header(text string) {
	minh.f.Header([]byte(text))
}

func (minh *BoundaryWriter) Block(cols []interface{}) {
	panic("Block() cannot be called for BoundaryWriter. Use")
}

func (minh *BoundaryWriter) Coordinates(x, y, z []float32) {
	minh.scaledBoundary = minh.boundary / minh.l

	coord := [3][]float32{ x, y, z }
	sizes := minh.cellSizes(coord)
	indices, boundaryFlag := minh.indices(coord, sizes)	

	minh.cellIndex = indices

	minh.boundaryColumn(boundaryFlag)
}

// indices returns the indices of the points in each cell+boundary region
// as well as the boundary flags for each region.
func (minh *BoundaryWriter) indices(
	coord [3][]float32, sizes []int,
) (indices [][]int, boundaryFlag [][]int8) {
	c := minh.cells
	dx := minh.l / float32(c)

	// Initialize buffers
	indices, boundaryFlag = make([][]int, c*c*c), make([][]int8, c*c*c)
	curr := make([]int, c*c*c)
	for i := range indices { 
		indices[i] = make([]int, sizes[i])
		boundaryFlag[i] = make([]int8, sizes[i])
	}

	// set boundaryFlag and index.
	update := func(vec [3]int, i int, flag int8) {
		g := gridIndex(vec, c)
		indices[g][curr[g]] = i
		boundaryFlag[g][curr[g]] = flag
		curr[g]++
	}

	// insert all points
	for i := range coord[0] {
		for k := 0; k < 3; k++ { minh.vec[k] = coord[k][i] / dx }
		minh.idxSum()
		update(minh.idx, i, 0)
		if minh.idx != minh.sum { update(minh.sum, i, 1) }
	}

	return indices, boundaryFlag
}

// gridIndex returns the index of vec within a grid of width cells.
func gridIndex(vec [3]int, cells int) int {
	return vec[0] + vec[1]*cells + vec[2]*cells*cells
}

// cellSizes returns the number of points in each cell.
func (minh *BoundaryWriter) cellSizes(coord [3][]float32) []int {
	c := minh.cells
	dx := minh.l / float32(c)	
	sizes := make([]int, c*c*c)

	// Increase size.
	update := func(vec [3]int) {
		g := gridIndex(minh.sum, c)
		sizes[g]++
	}

	// insert all points
	for i := range coord[0] {
		for k := 0; k < 3; k++ { minh.vec[k] = coord[k][i] / dx }
		minh.idxSum()
		update(minh.idx)
		if minh.idx != minh.sum { update(minh.sum) }
	}
	return sizes
}

// idxSum writes the index and sum = index + region for a given vector.
func (minh *BoundaryWriter) idxSum() {
	for k := 0; k < 3; k++ {
		minh.idx[k] = int(minh.vec[k])
		if minh.idx[k] >= minh.cells {
			minh.idx[k] -= minh.cells
			minh.vec[k] -= minh.l
		}
		reg := minh.region(minh.idx[k], minh.vec[k])

		minh.sum[k] = minh.idx[k] + reg
		if minh.sum[k] < 0 { minh.sum[k] += minh.cells }
		if minh.sum[k] >= minh.cells { minh.sum[k] -= minh.cells }
	}
}

// Region returns an int representing the location of x within cell ix. x needs
// to have already been scaled by dx. -1 indicates that x is within the boundary
// region of cell ix-1, +1 indicates that x is within the boundary region of
// cell ix+1, and 0 indicates that the point isn't in any cell's boundary
// region.
func (minh *BoundaryWriter) region(ix int, x float32) int {
	low := float32(ix)
	high := low + 1

	bLow := low + minh.scaledBoundary
	if x < bLow { return -1 }
	bHigh := high - minh.scaledBoundary
	if x > bHigh { return +1 }
	return 0
}

// Column writes a column with the given name, type information, and data
// to the BoundaryWriter. This column is split up into cells and boundaries.
func (minh *BoundaryWriter) Column(name string, col Column, x interface{}) {
	minh.cols = append(minh.cols, col)
	minh.names = append(minh.names, name)
	
	c := minh.cells

	for i := 0; i < c*c*c; i++ {
		idx := minh.cellIndex[i]
		N := len(idx)

		switch col.Type {
		case Int64, Int:
			minh.i64Buf = expandInt64(minh.i64Buf, N)
			buf := minh.i64Buf
			ix := x.([]int64)
			for j := range idx { buf[j] = ix[idx[j]] }

			if col.Type == Int64 {
				minh.f.FixedSizeGroup(minnow.Int64Group, N)
			} else {
				minh.f.IntGroup(N)
			}
			minh.f.Data(minh.i64Buf)
		case Float32, Float:
			minh.f32Buf = expandFloat32(minh.f32Buf, N)
			buf := minh.f32Buf
			fx := x.([]float32)
			for j := range idx { buf[j] = fx[idx[j]] }

			if col.Type == Float32 {
				minh.f.FixedSizeGroup(minnow.Float32Group, N)
			} else {
				processFloatGroup(minh.f32Buf, col)
			}
			minh.f.Data(minh.f32Buf)
		default:
			panic(fmt.Sprintf("Can't write column with type flag %d", col.Type))
		}
	}	
}

func (minh *BoundaryWriter) boundaryColumn(boundaryFlag [][]int8) {
	minh.cols = append(minh.cols, Column{ Type: Int })
	minh.names = append(minh.names, "boundary")

	c := minh.cells

	for i := 0; i < c*c*c; i++ {
		N := len(boundaryFlag[i])
		minh.i64Buf = expandInt64(minh.i64Buf, N)
		for j := range boundaryFlag[i] {
			minh.i64Buf[j] = int64(boundaryFlag[i][j])
			minh.f.FixedSizeGroup(minnow.IntGroup, N)
			minh.f.Data(minh.i64Buf)
		}
	}
}

// Close finalizes and closes the BoundaryWriter.
func (minh *BoundaryWriter) Close() {
	minh.f.Header([]byte(strings.Join(minh.names, "$")))
	minh.f.Header(minh.cols)
	minh.f.Header(geometry{ minh.l, minh.boundary, int64(minh.cells) })
	minh.f.Header(int64(minh.blocks))
	minh.f.Header(minh.blockSizes)
	minh.f.Close()
}
