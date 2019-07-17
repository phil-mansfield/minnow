package snapshot

import (
	"fmt"
	"runtime"
)

// Grid manages the geometry of a cube which has been split up into cubic
// segments. It should be embedded in a struct which contains cube data of a
// specific type.
type Grid struct {
	NCell int64 // Number of cells on one side of the grid.
	NSide int64 // Number of elements on one side of a cell.
}


// Index returns the locaiton of a given ID in a Grid. Here, the id is assumed
// to be in the form of ix + iy*cells + iz*cells^2, and the location is
// specified by two indices, c and i. c is the index of the cell that the ID
// is within, and i is the index within that cell.
func (g *Grid) Index(id int64) (c, i int64) {
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

// VectorGrid is a segmented cubic grid that stores float32 vectors in cubic
// sub-segments.
type VectorGrid struct {
	Grid
	Cells [][][3]float32
}

// NewVectorGrid creates a new VectorGrid made with the specified total side
// length and number of cells on one side. cells must cleanly divide nSideTot.
func NewVectorGrid(cells, nSideTot int) *VectorGrid {
	nSide := nSideTot / cells
	if nSide*cells != nSideTot {
		panic(fmt.Sprintf("cells = %d doesn't evenly divide nSideTot = %d.",
			cells, nSideTot))
	}

	vg := &VectorGrid{
		Cells: make([][][3]float32, cells*cells*cells),
	}
	vg.Grid = Grid{NCell: int64(cells), NSide: int64(nSide)}

	for i := range vg.Cells {
		vg.Cells[i] = make([][3]float32, nSide*nSide*nSide)
	}

	return vg
}

// XGrid creates a VectorGrid of the position vectors in a snapshot.
func XGrid(snap Snapshot, cells int) (*VectorGrid, error) {
	hd := snap.Header()
	files := snap.Files()

	grid := NewVectorGrid(cells, int(hd.NSide))

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

// VGrid creates a VectorGrid of the velocity vectors in a snapshot.
func VGrid(snap Snapshot, cells int) (*VectorGrid, error) {
	hd := snap.Header()
	files := snap.Files()

	grid := NewVectorGrid(cells, int(hd.NSide))

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

// Insert inserts a vector into a VectorGrid.
func (vg *VectorGrid) Insert(id int64, v [3]float32) {
	c, i := vg.Index(id)
	vg.Cells[c][i] = v
}

func (vg *VectorGrid) IntBuffer() [3][]uint64 {
	out := [3][]uint64{}
	for i := 0; i < 3; i++ {
		out[i] = make([]uint64, vg.NSide*vg.NSide*vg.NSide)
	}

	return out
}
