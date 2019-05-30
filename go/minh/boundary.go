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
	cellIndex [][]int
	i64Buf []int64
	f32Buf []float32
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
	c := minh.cells
	dx := minh.l / float32(c)

	lims := make([]float32, c+1)
	for i := range lims { lims[i] = float32(i)*dx }
	
	sizes := minh.cellSizes(x, y, z, lims)
	indices, boundaryFlag := minh.indices(x, y, z, lims, sizes)	

	minh.cellIndex = indices
	minh.Column("boundary", Column{Type: Int}, boundaryFlag)
}

func (minh *BoundaryWriter) indices(
	x, y, z, lims []float32, sizes []int,
) (indices [][]int, boundaryFlag []int64) {
	c, n := minh.cells, len(x)

	indices, boundaryFlag = make([][]int, c*c*c), make([]int64, c*c*c)
	for i := range indices { indices[i] = make([]int, sizes[i]) }
	curr := make([]int, c*c*c)
	for i := 1; i < len(curr); i++ {
		curr[i] = curr[i - 1] + sizes[i - 1]
	}

	idx := [3]int{ }
	sum := [3]int{ }
	coord := [3][]float32{ x, y, z }
	dx := minh.l / float32(c)

	for i := 0; i < n; i++ {
		for k := 0; k < 3; k++ {
			idx[k] = int(coord[k][i] / dx)
			reg := minh.region(idx[k], coord[k][i], lims)
			sum[k] = idx[k] + reg
			if sum[k] < 0 { sum[k] += c }
			if sum[k] >= c { sum[k] -= c }
		}

		g := idx[0] + c*idx[1] + c*c*idx[2]
		indices[g][curr[g]] = i
		boundaryFlag[curr[g]] = 0
		curr[g]++
		
		if idx[0] != sum[0] || idx[1] != sum[1] || idx[2] != sum[2] {
			g := sum[0] + c*sum[2] + c*c*sum[2]
			indices[g][curr[g]] = i
			boundaryFlag[curr[g]] = 1
			curr[g]++
		}
	}

	return indices, boundaryFlag
}

func (minh *BoundaryWriter) cellSizes(x, y, z, lims []float32) []int {
	c, n := minh.cells, len(x)
	idx := [3]int{ }
	sum := [3]int{ }
	coord := [3][]float32{ x, y, z }
	dx := minh.l / float32(c)	

	sizes := make([]int, c*c*c)

	for i := 0; i < n; i++ {
		for k := 0; k < 3; k++ {
			idx[k] = int(coord[k][i] / dx)
			reg := minh.region(idx[k], coord[k][i], lims)
			sum[k] = idx[k] + reg
			if sum[k] < 0 { sum[k] += c }
			if sum[k] >= c { sum[k] -= c }
		}

		g := idx[0] + c*idx[1] + c*c*idx[2]
		sizes[g]++
		
		if idx[0] != sum[0] || idx[1] != sum[1] || idx[2] != sum[2] {
			g := sum[0] + c*sum[2] + c*c*sum[2]
			sizes[g]++
		}
	}
	return sizes
}

func (minh *BoundaryWriter) region(ix int, x float32, lims []float32) int {
	low, high := lims[ix] + minh.boundary, lims[ix + 1] - minh.boundary
	if x < low { return -1 }
	if x > high { return +1 }
	return 0
}

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
			minh.f.Data(ix)
		case Float32, Float:
			minh.f32Buf = expandFloat32(minh.f32Buf, N)
			buf := minh.f32Buf
			fx := x.([]float32)
			for j := range idx { buf[j] = fx[idx[j]] }

			if col.Type == Float32 {
				minh.f.FixedSizeGroup(minnow.Float32Group, N)
			} else {
				processFloatGroup(fx, col)
			}
			minh.f.Data(fx)
		default:
			panic(fmt.Sprintf("Can't write column with type flag %d", col.Type))
		}
	}	
}

func (minh *BoundaryWriter) Close() {
	minh.f.Header([]byte(strings.Join(minh.names, "$")))
	minh.f.Header(minh.cols)
	minh.f.Header(geometry{ minh.l, minh.boundary, int64(minh.cells) })
	minh.f.Header(int64(minh.blocks))
	minh.f.Header(minh.blockSizes)
	minh.f.Close()
}
