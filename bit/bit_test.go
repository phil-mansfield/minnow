package bit

import (
	"os"
	"math/rand"
	"testing"
)

func TestArray(t *testing.T) {
	data := make([]uint64, 123)
	out := make([]uint64, len(data))
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	for bits := 1; bits <=64; bits++ {
		arr := NewArray(bits, data)
		arr.Slice(out)

		mask := ^(^uint64(0) << uint(bits))

		for i := range out {
			if out[i] != data[i] & mask {
				t.Errorf(
					"data[%d] & %x = %x, but out[%d] = %x",
					i, mask, data[i] & mask, i, out[i],
				)
			}
		}
	}
}

func TestArrayBuffer(t *testing.T) {
	fname := "../test_files/array_buffer.test"
	f, err := os.Create(fname)
	if err != nil { panic(err.Error()) }

	lengths := []int{10, 5, 1, 20}
	bits := make([]int, 4)
	ab := &ArrayBuffer{ }

	for i := range lengths {
		data := ab.Uint64(lengths[i])
		for j := range data { data[j] = uint64(j) }
		bits[i] = ab.Write(f, data)
	}

	f.Close()

	ab = &ArrayBuffer{ }

	f, err = os.Open(fname)
	if err != nil { panic(err.Error()) }

	for i := range lengths {
		data := ab.Read(f, bits[i], lengths[i])
		if len(data) != lengths[i] {
			t.Errorf("Expected len(array_%d) = %d, but got %d.", 
				i, lengths[i], len(data))
		}
		for j := range data {
			if int(data[j]) != j {
				t.Errorf("Expected array_%d[%d] = %d, but got %d.",
					i, j, j, data[j])
			}
		}
	}
}
