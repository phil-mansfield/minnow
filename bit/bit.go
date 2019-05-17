package bit

import (
	"fmt"
	"math"
)

// Array is an array in in which elements are packed with a width of
// b < 64 bits. It allows for space-efficient storage when integers have
// well-knownvalue ranges that don't correspond to exactly 64, 32, 16, or 8
// bits.
type Array struct {
	Length int
	Bits byte
	Data []byte
}

func ArrayBytes(bits, length int) int {
	return int(math.Ceil(float64(bits * length) / 8))
}

// Slice converts the contents of a Array into a standard uint64 slice.
// len(out) must equal arr.Length.
func (arr *Array) Slice(out []uint64) {
	if len(out) < arr.Length {
		panic(fmt.Sprintf("Array has length %d, but out buffer has " +
			"length %d.", arr.Length, len(out)))
	}

	// Set up buffers and commonly-used values.
	bits := int(arr.Bits)
	buf, tBuf := [8]byte{ }, [9]byte{ }
	bufBytes := uint64(arr.Bits / 8)
	if bufBytes * 8 < uint64(arr.Bits) { bufBytes++ }

	for i := 0; i < arr.Length; i++ {
		// Find where we are in the array.
		startBit := uint64(i*bits % 8)
		nextStartBit := (startBit + uint64(bits)) % 8

		startByte := int(i*bits / 8)
		endByte := int(((i + 1)*bits - 1) / 8)
		tBufBytes := endByte - startByte + 1

		// Pull bytes out into a buffer.
		for j := 0; j < tBufBytes; j++ {
			tBuf[j] = arr.Data[startByte + j]
		}

		// Mask unrelated edges
		startMask := (^byte(0)) << startBit
		endMask := (^byte(0)) >> (8 - nextStartBit)
		if nextStartBit == 0 { endMask = ^byte(0) }
		
		tBuf[0] &= startMask
		tBuf[tBufBytes - 1] &= endMask

		// Transfer shifted bytes into unshifted buffer.
		for j := uint64(0); j < bufBytes; j++ {
			buf[j] = tBuf[j] >> startBit
		}
		for j := uint64(0); j < bufBytes; j++ {
			buf[j] |= tBuf[j+1] << (8-startBit)
			
		}

		// Clear tBuf for next loop.
		for i := 0; i < tBufBytes; i++ { tBuf[i] = 0 }

		// Convert to uint64
		xi := uint64(0)
		for j := uint64(0); j < bufBytes; j++ {
			xi |= uint64(buf[j]) << (8*j)
		}
		out[i] = xi
	}
}

// NewArray creates a new Array which stores only the bits least
// signiticant bits of every element in x.
func NewArray(bits int, x []uint64) *Array {
	if bits > 64 {
		panic("Cannot pack more than 64 bits per element into a bit.Array")
	}

	// Set up buffers and commonly used values.
	nBytes := ArrayBytes(bits, len(x))
	arr := &Array{
		Length: len(x), Bits: byte(bits), Data: make([]byte, nBytes),
	}

	buf, tBuf := [8]byte{ }, [9]byte{ }
	bufBytes := uint64(bits / 8)
	if bufBytes * 8 < uint64(bits) { bufBytes++ }

	mask := (^uint64(0)) >> uint64(64 - bits)

	for i, xi := range x {
		xi &= mask
		currBit := uint64(i*bits % 8)

		// Move to byte-wise buffer.
		for j := uint64(0); j < bufBytes; j++ {
			buf[j] = byte(xi >> (8*j))
		}

		// Shift and move to the transfer buffer
		tBuf[bufBytes] = 0
		for j := uint64(0); j < bufBytes; j++ {
			tBuf[j] = buf[j] << currBit
		}
		for j := uint64(0); j < bufBytes; j++ {
			tBuf[j + 1] |= buf[j] >> (8-currBit)
		}

		// Transfer bits into the Array
		startByte := i * bits / 8
		endByte := ((i + 1)*bits - 1) / 8

		for j := 0; j < (endByte - startByte) + 1; j++ {
			arr.Data[startByte + j] |= tBuf[j]
		}
	}

	return arr
}
