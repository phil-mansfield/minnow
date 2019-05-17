package bit

import (
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

func BenchmarkArray10(b *testing.B) {
	bits := 10
	data := make([]uint64, 512)
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	b.SetBytes(int64(len(data) * bits))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewArray(bits, data)
	}
}

func BenchmarkArray20(b *testing.B) {
	bits := 20
	data := make([]uint64, 512)
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	b.SetBytes(int64(len(data) * bits))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewArray(bits, data)
	}
}

func BenchmarkArray40(b *testing.B) {
	bits := 40
	data := make([]uint64, 512)
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	b.SetBytes(int64(len(data) * bits))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewArray(bits, data)
	}
}

func BenchmarkArraySlice10(b *testing.B) {
	bits := 10
	data := make([]uint64, 512)
	out := make([]uint64, len(data))
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	arr := NewArray(bits, data)

	b.SetBytes(int64(len(data) * bits))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		arr.Slice(out)
	}
}


func BenchmarkArraySlice20(b *testing.B) {
	bits := 20
	data := make([]uint64, 512)
	out := make([]uint64, len(data))
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	arr := NewArray(bits, data)

	b.SetBytes(int64(len(data) * bits))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		arr.Slice(out)
	}
}

func BenchmarkArraySlice40(b *testing.B) {
	bits := 40
	data := make([]uint64, 512)
	out := make([]uint64, len(data))
	for i := range data {
		data[i] = uint64(rand.Int63())
	}

	arr := NewArray(bits, data)

	b.SetBytes(int64(len(data) * bits))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		arr.Slice(out)
	}
}
