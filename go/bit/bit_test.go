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
		bits[i] = ab.Bits(data)
		ab.Write(f, data, bits[i])
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

func benchmarkReadArrayN(b *testing.B, bits int) {
	x := make([]uint64, 100 * 1000)
	for i := range x { x[i] = uint64(i % 100) }
	buf := make([]byte, ArrayBytes(bits, len(x)))

	b.SetBytes(int64(8*len(x)))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		BufferedArray(bits, x, buf)
	}
}

func benchmarkWriteArrayN(b *testing.B, bits int) {
	x := make([]uint64, 100 * 1000)
	for i := range x { x[i] = uint64(i % 100) }
	arr := NewArray(bits, x)

	b.SetBytes(int64(8*len(x)))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		arr.Slice(x)
	}
}


func BenchmarkReadArray64(b *testing.B) { benchmarkReadArrayN(b, 64) }
func BenchmarkReadArray45(b *testing.B) { benchmarkReadArrayN(b, 45) }
func BenchmarkReadArray32(b *testing.B) { benchmarkReadArrayN(b, 32) }
func BenchmarkReadArray23(b *testing.B) { benchmarkReadArrayN(b, 21) }
func BenchmarkReadArray16(b *testing.B) { benchmarkReadArrayN(b, 16) }
func BenchmarkReadArray11(b *testing.B) { benchmarkReadArrayN(b, 11) }
func BenchmarkReadArray8(b *testing.B) { benchmarkReadArrayN(b, 8) }

func BenchmarkWriteArray64(b *testing.B) { benchmarkWriteArrayN(b, 64) }
func BenchmarkWriteArray45(b *testing.B) { benchmarkWriteArrayN(b, 45) }
func BenchmarkWriteArray32(b *testing.B) { benchmarkWriteArrayN(b, 32) }
func BenchmarkWriteArray23(b *testing.B) { benchmarkWriteArrayN(b, 21) }
func BenchmarkWriteArray16(b *testing.B) { benchmarkWriteArrayN(b, 16) }
func BenchmarkWriteArray11(b *testing.B) { benchmarkWriteArrayN(b, 11) }
func BenchmarkWriteArray8(b *testing.B) { benchmarkWriteArrayN(b, 8) }
