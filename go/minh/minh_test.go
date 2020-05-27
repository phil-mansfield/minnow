package minh

import (
	"math"
	"testing"

	minnow "github.com/phil-mansfield/minnow/go"
)

func TestReaderWriter(t *testing.T) {
	fname := "../../test_files/reader_writer_minh.test"
	
	names := []string{
		"int64", "float32", "int", "float", "log",
	}

	text := "Cats are the best. Don't we love them?!@#$%^&*(),.." + 
		"..[]{};':\"|\\/-=_+`~meow meow meow"
	L, boundary, cells := float32(100), float32(10), 4

	columns := []Column{
		Column{Type: minnow.Int64Group},
		Column{Type: minnow.Float32Group},
		Column{Type: minnow.IntGroup},
		Column{Type: minnow.FloatGroup, Low: 100, High: 200, Dx: 1},
		Column{Type: minnow.FloatGroup, Log: 1, Low: 10, High: 14, Dx: 0.01},
	}

	block1 := []interface{}{
		[]int64{100, 200, 300, 400, 500},
		[]float32{150, 250, 350, 450, 550},
		[]int64{-30, -35, -25, -10, -20},
		[]float32{100, 200, 125, 150, 100},
		[]float32{1e10, 1e11, 1e11, 1e14, 3e13},
	}

	block2 := []interface{}{
		[]int64{125, 225, 325},
		[]float32{1750, 2750, 3750},
		[]int64{1000, 1000, 1000},
		[]float32{100, 100, 100},
		[]float32{1e14, 1e14, 1e14},
	}

	joinedBlocks := []interface{}{
		[]int64{100, 200, 300, 400, 500, 125, 225, 325},
		[]float32{150, 250, 350, 450, 550, 1750, 2750, 3750},
		[]int64{-30, -35, -25, -10, -20, 1000, 1000, 1000},
		[]float32{100, 200, 125, 150, 100, 100, 100, 100},
		[]float32{1e10, 1e11, 1e11, 1e14, 3e13, 1e14, 1e14, 1e14},
	}

	blocks := [][]interface{}{ block1, block2}

	wr := Create(fname)
	wr.Header(names, text, columns)
	wr.Geometry(L, boundary, cells)
	for _, block := range blocks { wr.Block(block) }
	wr.Close()

	floatOut := map[string][]float32{ "float32": nil, "float": nil, "log": nil }
	intOut := map[string][]int64{ "int64": nil, "int": nil }

	blocks = append(blocks, joinedBlocks)

	rd := Open(fname)

	if !stringsEq(rd.Names, names) || rd.Text != text ||
		!columnsEq(rd.Columns, columns) || rd.Blocks != 2 ||
		rd.Length != 8 || !intsEq(rd.BlockLengths, []int{5, 3}) ||
		rd.L != 100.0 || rd.Boundary != 10.0 || rd.Cells != 4 {
		t.Fatalf("Could not read header: %v %v %v %v %v %v %v %v %v",
			stringsEq(rd.Names, names),
			rd.Text == text, columnsEq(rd.Columns, columns), rd.Blocks == 2,
			rd.Length == 8, intsEq(rd.BlockLengths, []int{5, 3}),
			rd.L == 100.0, rd.Boundary == 10.0, rd.Cells == 4)
	}

	for b, block := range blocks {
		if b < 2 {
			rd.IntBlock(b, intOut)
			rd.FloatBlock(b, floatOut)
		} else {
			intOut = rd.Ints([]string{"int64", "int"})
			floatOut = rd.Floats([]string{"float32", "float", "log"})
		}

		int64Col := block[0].([]int64)
		float32Col := block[1].([]float32)
		intCol := block[2].([]int64)
		floatCol := block[3].([]float32)
		logCol := block[4].([]float32)

		if !int64sEq(int64Col, intOut["int64"]) {
			t.Errorf("Block %d) read %v instead of %v for column '%s'.",
				b, intOut["int64"], int64Col, "int64")
		}
		if !float32sEq(float32Col, floatOut["float32"], 1e-6) {
			t.Errorf("Block %d) read %v instead of %v for column '%s'.",
				b, floatOut["float32"], float32Col, "float32")
		}
		if !int64sEq(intCol, intOut["int"]) {
			t.Errorf("Block %d) read %v instead of %v for column '%s'.",
				b, intOut["int"], intCol, "int")
		}
		if !float32sEq(floatCol, floatOut["float"], columns[3].Dx) {
			t.Errorf("Block %d) read %v instead of %v for column '%s'.",
				b, floatOut["float"], floatCol, "float")
		}
		if !log32sEq(logCol, floatOut["log"], columns[4].Dx) {
			t.Errorf("Block %d) read %v instead of %v for column '%s'.",
				b, floatOut["log"], logCol, "log")
		}
	}

	rd.Close()
}


func TestBoundaryRegion(t *testing.T) {
	L := float32(90.0)
	Bnd := float32(10.0)
	Cells := 3
	minh := &BoundaryWriter{ 
		Writer: Writer{ l: L, boundary: Bnd, cells: Cells },
		scaledBoundary: Bnd/L * float32(Cells),
	}

	tests := []struct {
		x float32
		exp int
	} {
		{0.0, -1},
		{0.5, 0},
		{0.9, +1},
		{1.0, -1},
		{1.5, 0},
		{1.9, +1},
		{2.0, -1},
		{2.5, 0},
		{2.9, +1},

	}

	for i := range tests {
		ix := int(tests[i].x)
		res := minh.region(ix, tests[i].x)
		if res != tests[i].exp {
			t.Errorf("%d) Expected region(%d, %f) = %d, but got %d",
				i, ix, tests[i].x, tests[i].exp, res)
		}
	}
}

func TestBoundaryIdxReg(t *testing.T) {
	L := float32(100.0)
	Cells := 2
	Bnd := float32(20.0)
	minh := &BoundaryWriter{ 
		Writer: Writer{ l: L, boundary: Bnd, cells: Cells },
		scaledBoundary: Bnd / L * float32(Cells),
	}

	tests := []struct{
		vec [3]float32
		expReg, expIdx [3]int
	} {
		{[3]float32{0.5, 0.5, 0.5}, [3]int{ 0,  0,  0}, [3]int{0, 0, 0}},
		{[3]float32{0.5, 0.5, 1.5}, [3]int{ 0,  0,  0}, [3]int{0, 0, 1}},
		{[3]float32{0.5, 1.5, 0.5}, [3]int{ 0,  0,  0}, [3]int{0, 1, 0}},
		{[3]float32{0.5, 1.5, 1.5}, [3]int{ 0,  0,  0}, [3]int{0, 1, 1}},
		{[3]float32{1.5, 0.5, 0.5}, [3]int{ 0,  0,  0}, [3]int{1, 0, 0}},
		{[3]float32{1.5, 0.5, 1.5}, [3]int{ 0,  0,  0}, [3]int{1, 0, 1}},
		{[3]float32{1.5, 1.5, 0.5}, [3]int{ 0,  0,  0}, [3]int{1, 1, 0}},
		{[3]float32{1.5, 1.5, 1.5}, [3]int{ 0,  0,  0}, [3]int{1, 1, 1}},

		{[3]float32{1.9, 1.5, 0.5}, [3]int{ 1,  0,  0}, [3]int{1, 1, 0}},
		{[3]float32{1.1, 1.5, 0.5}, [3]int{-1,  0,  0}, [3]int{1, 1, 0}},
		{[3]float32{1.5, 1.9, 0.5}, [3]int{ 0,  1,  0}, [3]int{1, 1, 0}},
		{[3]float32{1.5, 1.1, 0.5}, [3]int{ 0, -1,  0}, [3]int{1, 1, 0}},
		{[3]float32{1.5, 1.5, 0.9}, [3]int{ 0,  0,  1}, [3]int{1, 1, 0}},
		{[3]float32{1.5, 1.5, 0.1}, [3]int{ 0,  0, -1}, [3]int{1, 1, 0}},

		{[3]float32{0.9, 1.9, 0.5}, [3]int{ 1,  1,  0}, [3]int{0, 1, 0}},
		{[3]float32{0.1, 1.9, 0.5}, [3]int{-1,  1,  0}, [3]int{0, 1, 0}},
		{[3]float32{0.9, 1.1, 0.5}, [3]int{ 1, -1,  0}, [3]int{0, 1, 0}},
		{[3]float32{0.1, 1.1, 0.5}, [3]int{-1, -1,  0}, [3]int{0, 1, 0}},
		{[3]float32{0.5, 1.9, 0.9}, [3]int{ 0,  1,  1}, [3]int{0, 1, 0}},
		{[3]float32{0.5, 1.1, 0.9}, [3]int{ 0, -1,  1}, [3]int{0, 1, 0}},
		{[3]float32{0.5, 1.9, 0.1}, [3]int{ 0,  1, -1}, [3]int{0, 1, 0}},
		{[3]float32{0.5, 1.1, 0.1}, [3]int{ 0, -1, -1}, [3]int{0, 1, 0}},
		{[3]float32{0.9, 1.5, 0.9}, [3]int{ 1,  0,  1}, [3]int{0, 1, 0}},
		{[3]float32{0.1, 1.5, 0.9}, [3]int{-1,  0,  1}, [3]int{0, 1, 0}},
		{[3]float32{0.9, 1.5, 0.1}, [3]int{ 1,  0, -1}, [3]int{0, 1, 0}},
		{[3]float32{0.1, 1.5, 0.1}, [3]int{-1,  0, -1}, [3]int{0, 1, 0}},

		{[3]float32{0.9, 1.9, 1.9}, [3]int{ 1,  1,  1}, [3]int{0, 1, 1}},
		{[3]float32{0.9, 1.9, 1.1}, [3]int{ 1,  1, -1}, [3]int{0, 1, 1}},
		{[3]float32{0.9, 1.1, 1.9}, [3]int{ 1, -1,  1}, [3]int{0, 1, 1}},
		{[3]float32{0.9, 1.1, 1.1}, [3]int{ 1, -1, -1}, [3]int{0, 1, 1}},
		{[3]float32{0.1, 1.9, 1.9}, [3]int{-1,  1,  1}, [3]int{0, 1, 1}},
		{[3]float32{0.1, 1.9, 1.1}, [3]int{-1,  1, -1}, [3]int{0, 1, 1}},
		{[3]float32{0.1, 1.1, 1.9}, [3]int{-1, -1,  1}, [3]int{0, 1, 1}},
		{[3]float32{0.1, 1.1, 1.1}, [3]int{-1, -1, -1}, [3]int{0, 1, 1}},

		{[3]float32{2, 2, 2}, [3]int{-1, -1, -1}, [3]int{0, 0, 0}},
	}

	for i := range tests {
		idx, reg := minh.idxReg(tests[i].vec)
		
		if idx != tests[i].expIdx {
			t.Errorf("%d) Expected idxReg(%.1f) -> idx = %d, but got %d",
				i, tests[i].vec, tests[i].expIdx, idx)
		}
		if reg != tests[i].expReg {
			t.Errorf("%d) Expected idxReg(%.1f) -> reg = %d, but got %d",
				i, tests[i].vec, tests[i].expReg, reg)
		}
	}
}

func TestBoundaryHostCells(t *testing.T) {
	L := float32(100.0)
	Cells := 3
	Bnd := float32(20.0)
	minh := &BoundaryWriter{ 
		Writer: Writer{ l: L, boundary: Bnd, cells: Cells },
		scaledBoundary: Bnd / L * float32(Cells),
		cellBuf: make([]int, 8),
	}

	tests := []struct {
		idx, reg [3]int
		cells []int
	} {
		{ [3]int{0, 0, 0}, [3]int{0, 0, 0}, []int{0} },
		{ [3]int{1, 0, 0}, [3]int{0, 0, 0}, []int{1} },
		{ [3]int{0, 1, 0}, [3]int{0, 0, 0}, []int{3} },
		{ [3]int{1, 1, 0}, [3]int{0, 0, 0}, []int{4} },
		{ [3]int{0, 0, 1}, [3]int{0, 0, 0}, []int{9} },
		{ [3]int{1, 0, 1}, [3]int{0, 0, 0}, []int{10} },
		{ [3]int{0, 1, 1}, [3]int{0, 0, 0}, []int{12} },
		{ [3]int{1, 1, 1}, [3]int{0, 0, 0}, []int{13} },

		{ [3]int{1, 1, 1}, [3]int{ 1, 0, 0}, []int{13, 14} },
		{ [3]int{1, 1, 1}, [3]int{-1, 0, 0}, []int{13, 12} },
		{ [3]int{1, 1, 1}, [3]int{ 0, 1, 0}, []int{13, 16} },
		{ [3]int{1, 1, 1}, [3]int{ 0,-1, 0}, []int{13, 10} },
		{ [3]int{1, 1, 1}, [3]int{ 0, 0, 1}, []int{13, 22} },
		{ [3]int{1, 1, 1}, [3]int{ 0, 0,-1}, []int{13,  4} },

		{ [3]int{0, 0, 0}, [3]int{ 1, 1, 0}, []int{0, 1, 3, 4} },
		{ [3]int{0, 0, 0}, [3]int{ 0, 1, 1}, []int{0, 3, 9, 12} },
		{ [3]int{0, 0, 0}, [3]int{-1,-1, 0}, []int{0, 2, 6, 8} },
		{ [3]int{0, 0, 0}, [3]int{ 0,-1,-1}, []int{0, 6, 18, 24} },

		{ [3]int{0, 0, 0}, [3]int{ 1, 1, 1}, []int{0, 1, 3, 4, 9, 10, 12, 13} },
	}

	for i := range tests {
		cells := minh.hostCells(tests[i].idx, tests[i].reg)
		if !intsEq(cells, tests[i].cells) {
			t.Errorf("%d) Expected BoundaryWriter.hostCells(%d %d) = " + 
				"%d, but got %d", i, tests[i].idx, tests[i].reg,
				tests[i].cells, cells,
			)
		}
	}
}

func TestCellSizes(t *testing.T) {
	L := float32(100.0)
	cells := 2

	tests := []struct {
		bnd float32
		coord [3][]float32
		sizes []int
	} {
		{ 0, [3][]float32{{}, {}, {}},
			[]int{0, 0, 0, 0, 0, 0, 0, 0} },

		{ 0, [3][]float32{{ 0}, { 0}, { 0}},
			[]int{1, 0, 0, 0, 0, 0, 0, 0} },
		{ 0, [3][]float32{{50}, { 0}, { 0}},
			[]int{0, 1, 0, 0, 0, 0, 0, 0} },
		{ 0, [3][]float32{{ 0}, {50}, { 0}},
			[]int{0, 0, 1, 0, 0, 0, 0, 0} },
		{ 0, [3][]float32{{50}, {50}, { 0}},
			[]int{0, 0, 0, 1, 0, 0, 0, 0} },
		{ 0, [3][]float32{{ 0}, { 0}, {50}},
			[]int{0, 0, 0, 0, 1, 0, 0, 0} },
		{ 0, [3][]float32{{50}, { 0}, {50}},
			[]int{0, 0, 0, 0, 0, 1, 0, 0} },
		{ 0, [3][]float32{{ 0}, {50}, {50}},
			[]int{0, 0, 0, 0, 0, 0, 1, 0} },
		{ 0, [3][]float32{{50}, {50}, {50}},
			[]int{0, 0, 0, 0, 0, 0, 0, 1} },

		{ 20, [3][]float32{{ 0}, { 0}, { 0}},
			[]int{1, 1, 1, 1, 1, 1, 1, 1} },
		{ 20, [3][]float32{{50}, {50}, {50}},
			[]int{1, 1, 1, 1, 1, 1, 1, 1} },
		{ 20, [3][]float32{{ 0}, {50}, { 0}},
			[]int{1, 1, 1, 1, 1, 1, 1, 1} },

		{ 20, [3][]float32{{25}, {25}, {25}},
			[]int{1, 0, 0, 0, 0, 0, 0, 0} },

		{ 20, [3][]float32{{ 0}, {25}, {25}},
			[]int{1, 1, 0, 0, 0, 0, 0, 0} },

		{ 20, [3][]float32{{50}, {50}, {25}},
			[]int{1, 1, 1, 1, 0, 0, 0, 0} },
		{ 20, [3][]float32{{50}, {25}, {25}},
			[]int{1, 1, 0, 0, 0, 0, 0, 0} },

	}

	for i := range tests {
		minh := BoundaryWriter{
			Writer: Writer{ l: L, boundary: tests[i].bnd, cells: cells },
			cellBuf: make([]int, 8),
			scaledBoundary: tests[i].bnd / L * float32(cells),
		}

		sizes := minh.cellSizes(tests[i].coord)
		if !intsEq(sizes, tests[i].sizes) {
			t.Errorf("%d) expected BoundaryWriter.cellSizes(%g) = %d, " +
				"but got %d.", i, tests[i].coord, tests[i].sizes, sizes)
		}
	}
}

func TestBoundary(t *testing.T) {
	fname := "../../test_files/boundary_minh.test"

	vecs := [][3]float32{
		{25, 25, 25},
		{50, 50, 50},
		{26, 26, 95},
	}
	blocks := []struct{
		x []float32
		boundaryFlag []int64
		id []int64
	} {
		{[]float32{25, 50, 26}, []int64{0, 1, 1}, []int64{0, 1, 2}},
		{[]float32{50}, []int64{1}, []int64{1}},
		{[]float32{50}, []int64{1}, []int64{1}},
		{[]float32{50}, []int64{1}, []int64{1}},
		{[]float32{50, 26}, []int64{1, 0}, []int64{1, 2}},
		{[]float32{50}, []int64{1}, []int64{1}},
		{[]float32{50}, []int64{1}, []int64{1}},
		{[]float32{50}, []int64{0}, []int64{1}},
	}

	coord := [3][]float32{
		make([]float32, len(vecs)),
		make([]float32, len(vecs)),
		make([]float32, len(vecs)),
	}
	for i := range vecs {
		for k := 0; k < 3; k++ { coord[k][i] = vecs[i][k] }
	}

	id := make([]int64, len(vecs))
	for i := range id { id[i] = int64(i) }

	f := CreateBoundary(fname)
	f.Header("This is my header string.")
	f.Geometry(100.0, 20.0, 2)
	f.Coordinates(coord[0], coord[1], coord[2])
	f.Column("id", Column{Type: Int64}, id)
	f.Column("x", Column{Type: Float32}, coord[0])
	f.Close()

	rd := Open(fname)

	iOut := map[string][]int64 { "boundary": nil, "id": nil }
	fOut := map[string][]float32 { "x": nil }
		_ = fOut

	for b := 0; b < 8; b++ {
		rd.IntBlock(b, iOut)
		rd.FloatBlock(b, fOut)

		if !int64sEq(iOut["boundary"], blocks[b].boundaryFlag) {
			t.Errorf("Expected boundary[%d] = %d, but got %d.", b,
				blocks[b].boundaryFlag, iOut["boundary"])
		}
		if !int64sEq(iOut["id"], blocks[b].id) {
			t.Errorf("Expected id[%d] = %d, but got %d.", b,
				blocks[b].id, iOut["id"])
		}
		if !float32sEq(fOut["x"], blocks[b].x, 0.1) {
			t.Errorf("Expected x[%d] = %g, but got %g.", b,
				blocks[b].x, fOut["x"])
		}
	}

	rd.Close()
}

func stringsEq(x, y []string) bool {
	if len(x) != len(y) { return false }
	for i := range x { if x[i] != y[i] { return false } }
	return true
}

func intsEq(x, y []int) bool {
	if len(x) != len(y) { return false }
	for i := range x { if x[i] != y[i] { return false } }
	return true
}

func columnsEq(x, y []Column) bool {
	if len(x) != len(y) { return false }
	for i := range x { if x[i] != y[i] { return false } }
	return true
}

func int64sEq(x, y []int64) bool {
	if len(x) != len(y) { return false }
	for i := range x { if x[i] != y[i] { return false } }
	return true
}

func float32sEq(x, y []float32, dx float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] + dx < y[i] || x[i] - dx > y[i] { return false }
	}
	return true
}

func log32sEq(x, y []float32, dx float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		lx := float32(math.Log10(float64(x[i])))
		ly := float32(math.Log10(float64(y[i])))
		if lx + dx < ly || lx - dx > ly { return false }
	}
	return true
}
