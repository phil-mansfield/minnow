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
		t.Fatalf("Could not read header: %v %v %v %v %v %v",
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


func BoundaryRegionTest(t *testing.T) {
	L := float32(90.0)
	Bnd := float32(10.0)
	Cells := 3
	lims := []float32{ 0, 30, 60, 90 }

	tests := []struct {
		x float32
		exp int
	} {
		{0, -1},
		{15, 0},
		{25, +1},
		{30, -1},
		{45, 0},
		{55, +1},
		{60, -1},
		{75, 0},
		{85, +1},

	}

	minh := &BoundaryWriter{ 
		Writer: Writer{ l: L, boundary: Bnd, cells: Cells },
	}

	for i := range tests {
		ix := int(tests[i].x / 30)
		res := minh.region(ix, tests[i].x, lims)
		if res != tests[i].exp {
			t.Errorf("%d) Expected region(%d, %f) = %d, but got %d",
				i, ix, tests[i].x, tests[i].exp, res)
		}
	}
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
