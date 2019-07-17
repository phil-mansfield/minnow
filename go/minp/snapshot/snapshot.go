package snapshot

import (
	"encoding/binary"
	"io"
	"reflect"

	"github.com/phil-mansfield/nbody-utils/cosmo"

	"unsafe"
)

type Snapshot interface {
	Files() int // Number of files in the snapshot
	Header() *Header // Header contains basic information about the snapshot
	RawHeader(i int) []byte // Return the bytes of the original header block.
	UpdateHeader(hd *Header) // Change the header to a new one.
	UniformMass() bool // True if all particles are the same mass.

	// All these methods return internal buffers, so don't append to them or
	// expect them to stick around after the function is called again.
	ReadX(i int) ([][3]float32, error) // Read positions for file i.
	ReadV(i int) ([][3]float32, error) // Read velocities for file i.
	ReadID(i int) ([]int64, error) // Read IDs for file i.
	ReadMp(i int) ([]float32, error) // Read particle masses for file i.
}

// Header is a struct containing basic information about the snapshot. Not all
// simulation headers provide all information: the user is responsible for
// supplying that information afterwards in these cases.
type Header struct {
	Z, Scale float64 // Redshift, scale factor
	OmegaM, OmegaL, H100 float64 // Omega_m(z=0), Omega_L(z=0), little-h(z=0)
	L, Epsilon float64 // Box size, force softening
	NSide, NTotal int64 // Particles on one size, total particles
	UniformMp float64 // If all particle masses are the same, this is m_p.
}


func (hd *Header) calcUniformMass() {
	rhoM0 := cosmo.RhoAverage(hd.H100*100, hd.OmegaM, hd.OmegaL, 0)
	mTot := (hd.L * hd.L * hd.L) * rhoM0
	hd.UniformMp =  mTot / float64(hd.NTotal)
}

func readVecAsByte(rd io.Reader, end binary.ByteOrder, buf [][3]float32) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 12
	hd.Cap *= 12

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen*3; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4+j, i*4+3-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 12
	hd.Cap /= 12

	return nil
}

func readInt64AsByte(rd io.Reader, end binary.ByteOrder, buf []int64) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 8
	hd.Cap *= 8

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 4; j++ {
				idx1, idx2 := i*8+j, i*8+7-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 8
	hd.Cap /= 8

	return nil
}

func readInt32AsByte(rd io.Reader, end binary.ByteOrder, buf []int32) error {
	bufLen := len(buf)

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 4
	hd.Cap *= 4

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4+j, i*4+3-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 4
	hd.Cap /= 4

	return nil
}

func readFloat32AsByte(rd io.Reader, end binary.ByteOrder, buf []float32) error {
	bufLen := len(buf)
	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Len *= 4
	hd.Cap *= 4

	byteBuf := *(*[]byte)(unsafe.Pointer(&hd))
	_, err := rd.Read(byteBuf)
	if err != nil {
		return err
	}

	if !IsSysOrder(end) {
		for i := 0; i < bufLen; i++ {
			for j := 0; j < 2; j++ {
				idx1, idx2 := i*4+j, i*4+3-j
				byteBuf[idx1], byteBuf[idx2] = byteBuf[idx2], byteBuf[idx1]
			}
		}
	}

	hd.Len /= 4
	hd.Cap /= 4

	return nil
}

func IsSysOrder(end binary.ByteOrder) bool {
	buf32 := []int32{1}

	hd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf32))
	hd.Len *= 4
	hd.Cap *= 4

	buf8 := *(*[]int8)(unsafe.Pointer(&hd))
	if buf8[0] == 1 {
		return binary.LittleEndian == end
	}
	return binary.BigEndian == end
}
