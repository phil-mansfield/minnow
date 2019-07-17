package snapshot

import (
	"fmt"
	"runtime"
)

// Grid is a wrapper around a Snapshot which transforms it into a 
// Lagrangian-contiguous grid.
type Grid struct {
	hd *Header
	snap Snapshot
	cells int
	xGrid, vGrid *vectorGrid

	idBuffer []int64
	mpBuffer []float32
}

func NewGrid(snap Snapshot, cells int) *Grid {
	g := &Grid{ snap: snap, cells: cells, hd: snap.Header() }

	if !snap.UniformMass() {
		panic("snapshot.Grid can only be creates from Snapshots with " +
			"unfirom mass.")
	} else if g.hd.NSide % int64(cells) != 0 {
		panic(fmt.Sprintf("Snapshot has NSide = %d, but %d cells were " + 
			"requested.", g.hd.NSide, cells))
	}

	return g
}

func (g *Grid) Files() int { return g.cells*g.cells*g.cells }
func (g *Grid) Header() *Header { return g.snap.Header() }
func (g *Grid) RawHeader(i int) []byte { return g.snap.RawHeader(i) }
func (g *Grid) UpdateHeader(hd *Header) { g.snap.UpdateHeader(hd) }
func (g *Grid) UniformMass() bool { return g.snap.UniformMass() }

func (g *Grid) ReadX(i int) ([][3]float32, error) {
	if xGrid == nil {
		var err error
		g.xGrid, err = xGrid(g.snap, g.cells)
		if err != nil { return nil, err }
	}

	return g.xGrid.Cells[i], nil
}

func (g *Grid) ReadV(i int) ([][3]float32, error) {
	if vGrid == nil {
		var err error
		g.vGrid, err = vGrid(g.snap, g.cells)
		if err != nil { return nil, err }
	}

	return g.xGrid.Cells[i], nil
}

func (g *Grid) ReadID(idx int) ([]int64, error) {
	nSide := g.hd.NSide
	nFile := nSide / int64(g.cells)	
	fx := int64(idx % g.cells)
	fy := int64((idx / g.cells) % g.cells)
	fz := int64(idx / (g.cells*g.cells))

	if g.idBuffer == nil {
		g.idBuffer = make([]int64, nFile*nFile*nFile)
	}
	out := g.idBuffer

	// i is the index within the whole simulation, j is the index within the
	// file's array.
	ix0, iy0, iz0 := int64(fx*nFile), int64(fy*nFile), int64(fz*nFile)
	j := 0
	for jz := int64(0); jz < nFile; jz++ {
		for jy := int64(0); jy < nFile; jy++ {
			for jx := int64(0); jx < nFile; jx++ {
				ix, iy, iz := jx+ix0, jy+iy0, jz+iz0
				i := ix + iy*nSide + iz*nSide*nSide
				out[j] = i
				j++
			}
		}
	}

	return out, nil
}

func (g *Grid) ReadMp(i int) ([]float32, error) {
	nFile := g.hd.NSide / int64(g.cells)
	if g.mpBuffer == nil {
		g.mpBuffer = make([]float32, nFile*nFile*nFile)
	}
	out := g.mpBuffer
	
	mp := float32(g.hd.UniformMp)
	for i := range out { out[i] = mp }

	return out, nil
}

// Grid manages the geometry of a cube which has been split up into cubic
// segments. It should be embedded in a struct which contains cube data of a
// specific type.
type grid struct {
	NCell int64 // Number of cells on one side of the grid.
	NSide int64 // Number of elements on one side of a cell.
}


// Index returns the locaiton of a given ID in a Grid. Here, the id is assumed
// to be in the form of ix + iy*cells + iz*cells^2, and the location is
// specified by two indices, c and i. c is the index of the cell that the ID
// is within, and i is the index within that cell.
func (g *grid) Index(id int64) (c, i int64) {
	nAll := g.NCell * g.NSide

	if id < 0 || id >= nAll*nAll*nAll {
		panic(fmt.Sprintf("ID %d is not valid for NCell = %d, NSide = %d",
			id, g.NCell, g.NSide))
	}

	idx := id % nAll
	idy := (id / nAll) % nAll
	idz := id / (nAll * nAll)

	ix, iy, iz := idx % g.NSide, idy % g.NSide, idz % g.NSide
	i = ix + iy*g.NSide + iz*g.NSide*g.NSide

	cx, cy, cz := idx / g.NSide, idy / g.NSide, idz / g.NSide
	c = cx + cy*g.NCell + cz*g.NCell*g.NCell

	return c, i
}

// vectorGrid is a segmented cubic grid that stores float32 vectors in cubic
// sub-segments.
type vectorGrid struct {
	grid
	Cells [][][3]float32
}

// NewVectorGrid creates a new vectorGrid made with the specified total side
// length and number of cells on one side. cells must cleanly divide nSideTot.
func newVectorGrid(cells, nSideTot int) *vectorGrid {
	nSide := nSideTot / cells
	if nSide*cells != nSideTot {
		panic(fmt.Sprintf("cells = %d doesn't evenly divide nSideTot = %d.",
			cells, nSideTot))
	}

	vg := &vectorGrid{
		Cells: make([][][3]float32, cells*cells*cells),
	}
	vg.grid = grid{NCell: int64(cells), NSide: int64(nSide)}

	for i := range vg.Cells {
		vg.Cells[i] = make([][3]float32, nSide*nSide*nSide)
	}

	return vg
}

// XGrid creates a vectorGrid of the position vectors in a snapshot.
func xGrid(snap Snapshot, cells int) (*vectorGrid, error) {
	hd := snap.Header()
	files := snap.Files()

	grid := newVectorGrid(cells, int(hd.NSide))

	for i := 0; i < files; i++ {
		runtime.GC()

		x, err := snap.ReadX(i)
		if err != nil { return nil, err }
		id, err := snap.ReadID(i)
		if err != nil { return nil, err }

		for j := range x { grid.Insert(id[j] - 1, x[j]) }
	}

	return grid, nil
}

// VGrid creates a vectorGrid of the velocity vectors in a snapshot.
func vGrid(snap Snapshot, cells int) (*vectorGrid, error) {
	hd := snap.Header()
	files := snap.Files()

	grid := newVectorGrid(cells, int(hd.NSide))

	for i := 0; i < files; i++ {
		runtime.GC()

		v, err := snap.ReadV(i)
		if err != nil { return nil, err }
		id, err := snap.ReadID(i)
		if err != nil { return nil, err }
		for j := range v { grid.Insert(id[j] - 1, v[j]) }
	}

	return grid, nil
}

// Insert inserts a vector into a vectorGrid.
func (vg *vectorGrid) Insert(id int64, v [3]float32) {
	c, i := vg.Index(id)
	vg.Cells[c][i] = v
}

func (vg *vectorGrid) IntBuffer() [3][]uint64 {
	out := [3][]uint64{}
	for i := 0; i < 3; i++ {
		out[i] = make([]uint64, vg.NSide*vg.NSide*vg.NSide)
	}

	return out
}
