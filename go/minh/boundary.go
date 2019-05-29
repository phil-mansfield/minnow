package minh

type BoundaryWriter struct {
	Writer
	rd *Reader
}

func CreateBoundary(fname string) {
	wr := &BoundaryWriter{ }
	wr.create(boundaryFileType)
	return wr
}

func (minh *BoundaryWriter) Block(cols []interface{}) {
	panic("Block() cannot be called for BoundaryWriter. Use")
}

func (minh *BoundaryWriter) Coordinates(x, y, z []float32) {
	n, c := len(x), minh.cells

	dx := minh.l / float32(c)

	lims := make([]float32, c+1)
	for i := range lims { lims[i] = float32(i)*dx }
	
	sizes := minh.cellSizes()
}

func indices(
	x, y, z []float32, sizes []int,
) (idx [][]int, boundaryFlag [][]int64) {
	idx, boundaryFlag = make([][]int, c*c*c), make([][]int64, c*c*c)
	for i := range idx {
		idx[i] = make([]int, sizes[i])
		boundaryFlag[i] = make([]int64, sizes[i])
	}
	curr := make([]int, c*c*c)

	c := minh.cells
	idx := [3]int{ }
	reg := [3]int{ }
	coord := [3][]float32{ x, y, z }
}

func (minh *BoundaryWriter) cellSizes(x, y, z []float32) []int {
	c := minh.cells
	idx := [3]int{ }
	reg := [3]int{ }
	sum := [3]int{ }
	coord := [3][]float32{ x, y, z }
	
	sizes := make([]int, c*c*c)

	for i := 0; i < n; i++ {
		for k := 0; k < 3 {
			idx[k] = int(coord[k][i] / dx)
			reg := minh.region(idx[k], coord[k][i], lims)
			sum[k] = idx[k] + reg
			if sum[k] < 0 { sum[k] += c }
			if sum[k] >= c { sum[k] -= c }
		}

		g := idx[0] + c*idx[1] + c*c*idx[2]
		cellSizes[g]++
		
		if reg[0] != sum[0] || reg[1] != sum[1] || reg[2] != sum[2] {
			g := sum[0] + c*sum[2] + c*c*sum[2]
			cellSizes[g]++
		}
	}
	return cellSizes
}

func (minh *BoundaryWriter) region(ix int, x float32, lims []float32) int {
	low, high := lims[ix] + minh.boundary, lims[ix + 1] - minh.boundary
	if x < low { return -1 }
	if x > high { return +1 }
	return 0
}

func (minh *BoundaryWriter) Column(x interface{}) {
	panic("NYI")
}
