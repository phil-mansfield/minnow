package text

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func fakeReader(text string, itemSize, blockSize int) *Reader {
	f := bytes.NewReader([]byte(text))
	rd := &Reader{
		f: f, closer: ioutil.NopCloser(f),
		config: DefaultReaderConfig,
	}
	rd.config.MaxItemSize = int64(itemSize)
	rd.config.MaxBlockSize = int64(blockSize)
	return rd
}

func TestLineHeader(t *testing.T) {
	tests := []struct {
		text string
		itemSize int
		lines int
		res string
	} {
		{"a\nb\nc\nd", 100, 2, "a\nb"},
		{"a\nb\nc\nd", 5, 2, "a\nb"},
	}

	for i := range tests {
		rd := fakeReader(tests[i].text, tests[i].itemSize, 1000)
		out := rd.LineHeader(tests[i].lines)

		if out != tests[i].res {
			t.Errorf("Test %d: expected '%s', got '%s'.", i, tests[i].res, out)
		}
	}
}

func TestCommentHeader(t *testing.T) {
	tests := []struct {
		text string
		itemSize int
		res string
	} {
		{"#a\n#b\nc\nd", 100, "#a\n#b"},
		{"#a\n#b\nc\nd", 8, "#a\n#b"},
	}

	for i := range tests {
		rd := fakeReader(tests[i].text, tests[i].itemSize, 1000)
		out := rd.CommentHeader()

		if out != tests[i].res {
			t.Errorf("Test %d: expected '%s', got '%s'.",
				i, tests[i].res, out)
		}
	}
}


func TestReader(t *testing.T) {
	text := []byte(`#123456789012345678
#123456789012345678
1    2     3      5
11  12    13     15
21  22    23     25
31  32    33     35
41  42    43     45
`)
	itemSize := 50
	blockSize := 120

	config := DefaultReaderConfig
	config.MaxItemSize = int64(itemSize)
	config.MaxBlockSize = int64(blockSize)

	names := []string{"1", "2", "3", "4"}
	out1 := []interface{}{ []float32{}, []int64{}, []float32{}, []int64{} }
	out2 := []interface{}{ []float32{}, []int64{}, []float32{}, []int64{} }
	
	exp1 := []interface{} {
		[]float32{5, 15},
		[]int64{1, 11},
		[]float32{2, 12},
		[]int64{3, 13},
	}
	exp2 := []interface{} {
		[]float32{25, 35, 45},
		[]int64{21, 31, 41},
		[]float32{22, 32, 42},
		[]int64{23, 33, 43},
	}

	f := openFromReader(bytes.NewReader(text), config)
	f.SetNames(names)

	if f.Blocks() != 2 {
		t.Errorf("Expected 2 blocks, go %d.", f.Blocks())
	}

	f.Block(0, []string{"4", "1", "2", "3"}, out1)
	f.Block(1, []string{"4", "1", "2", "3"}, out2)

	for i := range exp1 {
		if !genericEq(exp1[i], out1[i], 1e-3) {
			t.Errorf("Expected %v for column %d of block %d, but got %v",
				exp1[i], i, 0, out1[i])
		}
	}
	for i := range exp2 {
		if !genericEq(exp2[i], out2[i], 1e-3) {
			t.Errorf("Expected %v for column %d of block %d, but got %v",
				exp2[i], i, 0, out2[i])
		}
	}
}

func genericEq(x, y interface{}, eps float32) bool {
	switch xs := x.(type) {
	case []int64:
		ys := y.([]int64)
		if len(xs) != len(ys) { return false }
		for i := range xs {
			if xs[i] != ys[i] { return false }
		}
		return true
	case []float32:
		ys := y.([]float32)
		if len(xs) != len(ys) { return false }
		for i := range xs {
			if xs[i] + eps < ys[i] || xs[i] - eps > ys[i] { return false }
		}
		return true
	}
	panic("Bad type")
}

func TestNextBlock(t *testing.T) {
	text := `1234
1234
1234
1234
1234
1234
`
	f := fakeReader(text, 6, 12)
	size := readerSize(f.f)

	for pos := int64(0); pos < size; pos++ {
		expected := pos + 12 - 6
		col := expected % 5
		expected += 5 - col

		if pos + 12 >= 30 { expected = -1 }

		f.f.Seek(pos, 0)
		next := f.nextBlock(size)
		pos2, _ := f.f.Seek(0, 1)

		if next != expected {
			t.Errorf("Expected next block = %d for pos = %d, but got %d",
				expected, pos, next)
		}

		if next != -1 && pos2 != next {
			t.Error("nextBlock did not set position to start of next block.")
		}
	}
}
