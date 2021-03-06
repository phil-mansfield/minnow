package minp

import (
	"fmt"
	"math"

	minnow "github.com/phil-mansfield/minnow/go"
)

const (
	Magic = 0xbadf00d
	Version = 0

	basicFileType int64 = iota
)

type idHeader struct {
	Magic, Version, FileType int64
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

type Cell struct {
	FileIndex, FileCells, SubCells int64
}

func (c *Cell) NFile(nSide int) int {
	if nSide < 0 || c.FileCells < 0 || nSide % int(c.FileCells) != 0 {
		panic(fmt.Sprintf("NSide = %d not a valid combination with " + 
			"FileCells = %d", nSide, c.FileCells))
	}
	return nSide / int(c.FileCells)
}

func (c *Cell) FileCoord() (fx, fy, fz int) {
	fx = int(c.FileIndex % c.FileCells)
	fy = int((c.FileIndex / c.FileCells) % c.FileCells)
	fz = int(c.FileIndex / (c.FileCells*c.FileCells))
	return fx, fy, fz
}

////////////
// Writer //
////////////

type Writer struct {
	f *minnow.Writer
	c Cell
	l float32
	hd Header
	periodic bool
	dx float32
}

func Create(fname string) *Writer {
	minp := &Writer{ }
	minp.f = minnow.Create(fname)
	minp.f.Header(idHeader{Magic, Version, basicFileType})
	return minp
}

func (minp *Writer) Header(
	hd *Header, rawHd []byte, c Cell, dx float64, periodic bool,
) {
	minp.f.Header(hd)
	minp.f.Header(rawHd)
	minp.f.Header(c)
	minp.f.Header(dx)
	minp.f.Header(boolToByte(periodic))

	minp.hd = *hd
	minp.c = c
	minp.periodic = periodic
	minp.dx = float32(dx)
}

func (minp *Writer) Vectors(vec [][3]float32) {
	var min, max [3]float32
	if minp.periodic {
		L := float32(minp.hd.L)
		min, max = [3]float32{ 0, 0, 0 }, [3]float32{ L, L, L }
	} else {
		min, max = bounds(vec)
		for i := range max {
			max[i] = math.Nextafter32(max[i], 2*max[i])
		}
	}

	nFile := minp.c.NFile(int(minp.hd.NSide))
	subCells := int(minp.c.SubCells)
	nSub := nFile / subCells
	nSub3, subCells3 := nSub*nSub*nSub, subCells*subCells*subCells

	if nFile*nFile*nFile != len(vec) {
		panic(fmt.Sprintf("len(vec) = %d, but NSide = %d and FileCells = %d",
			len(vec), minp.hd.NSide, minp.c.FileCells))
	}

	subBuf := [3][]float32{
		make([]float32, nSub3), make([]float32, nSub3), make([]float32, nSub3),
	}

	for k := 0; k < 3; k++ {
		minp.f.FloatGroup(nSub3, [2]float32{min[k], max[k]}, minp.dx)
		for sc := 0; sc < subCells3; sc++ {
			getSubCell(vec, subBuf, sc, subCells, nSub)
			minp.f.Data(subBuf[k])
		}
	}
}

func (minp *Writer) Close() {
	minp.f.Close()
}

////////////
// Reader //
////////////

// Reader contains methods for reading minp files.
type Reader struct {
	Header
	RawHeader []byte
	FileIndex, FileCells int
	Dx float64
	Periodic bool

	c Cell
	f *minnow.Reader
}

// Open opens a minp file with the given file name.
func Open(fname string) *Reader {
	minp := &Reader{ }
	minp.f = minnow.Open(fname)

	idHeader := idHeader{ }
	minp.f.Header(0, &idHeader)
	if idHeader.Magic != Magic {
		panic(fmt.Sprintf("Not a minp file. Magic number is %d, not %d",
			idHeader.Magic, Magic))
	} else if idHeader.Version != Version {
		panic(fmt.Sprintf("File version = %d, but code version = %d.",
			idHeader.Version, Version))
	} else if idHeader.FileType != basicFileType {
		panic(fmt.Sprintf("File type = %d", idHeader.FileType))
	}

	minp.f.Header(1, &minp.Header)
	minp.RawHeader = make([]byte, minp.f.HeaderSize(2))
	minp.f.Header(2, minp.RawHeader)
	minp.f.Header(3, &minp.c)
	minp.f.Header(4, &minp.Dx)
	bytePeriodic := byte(0)
	minp.f.Header(5, &bytePeriodic)
	minp.Periodic = byteToBool(bytePeriodic)

	minp.FileIndex = int(minp.c.FileIndex)
	minp.FileCells = int(minp.c.FileCells)

	return minp
}

// Vec reads the vectors form the file into out. out should have length equal to
// minp.N().
func (minp *Reader) Vectors(out [][3]float32) {
	nFile := minp.c.NFile(int(minp.NSide))
	subCells := int(minp.c.SubCells)
	nSub := nFile / subCells
	L := float32(minp.L)

	subCells3, nSub3 := subCells*subCells*subCells, nSub*nSub*nSub
	subBuf := [3][]float32{
		make([]float32, nSub3), make([]float32, nSub3), make([]float32, nSub3),
	}

	if minp.f.Blocks() != 3*subCells3 {
		panic(fmt.Sprintf("Expected %d sub-cells, but got %d",
			3*subCells, minp.f.Blocks()))
	}

	for sc := 0; sc < subCells3; sc++ {
		for k := 0; k < 3; k++ {
			minp.f.Data(k*subCells3 + sc, subBuf[k])

			if minp.Periodic {
				for i, x := range subBuf[k] {
					if x < 0 {
						subBuf[k][i] = x + L
					} else if x >= L {
						subBuf[k][i] = x - L
					}
				}
			}
		}
		setSubCell(out, subBuf, sc, subCells, nSub)
	}
}

// ID returns the Lagrangian IDs of the particles in the file.
func (minp *Reader) IDs(out []int64) {
	nFile := int64(minp.c.NFile(int(minp.NSide)))
	nSide := int64(minp.NSide)
	ifx, ify, ifz := minp.c.FileCoord()
	fx, fy, fz := int64(ifx), int64(ify), int64(ifz)
	
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
}

// N returns the number of particles in the file.
func (minp *Reader) N() int {
	return minp.f.Blocks() / 3
}

// Close closes the Reader.
func (minp *Reader) Close() {
	minp.f.Close()
}

// getSubCell sets subBuf with the corresponding values in x. x is a large
// vector array, subBuf is a set of small buffers corresponding to one sub-cell,
// sc is the index of the subcell in x, subCells is the number of sub-cells in
// x, and nSub is the length of one side of sub-cell
func getSubCell(x [][3]float32, subBuf [3][]float32, sc, subCells, nSub int) {
	nFile := nSub * subCells
	sx := sc % subCells
	sy := (sc / subCells) % subCells
	sz := sc / (subCells*subCells)

	ix0, iy0, iz0 := nSub*sx, nSub*sy, nSub*sz
	j := 0
	for jz := 0; jz < nSub; jz++ {
		for jy := 0; jy < nSub; jy++ {
			for jx := 0; jx < nSub; jx++ {
				ix, iy, iz := jx+ix0, jy+iy0, jz+iz0
				i := ix + iy*nFile + iz*nFile*nFile
				for k := 0; k < 3; k++ { subBuf[k][j] = x[i][k] }
				j++
			}
		}
	}
}

// getSubCell sets the corresponding values in x with the values of subBuf. x
// is a large vector array, subBuf is a set of small buffers corresponding to
// one sub-cell, sc is the index of the subcell in x, subCells is the number of
// sub-cells in x, and nSub is the length of one side of sub-cell
func setSubCell(x [][3]float32, subBuf [3][]float32, sc, subCells, nSub int) {
	nFile := nSub * subCells
	sx := sc % subCells
	sy := (sc / subCells) % subCells
	sz := sc / (subCells*subCells)

	ix0, iy0, iz0 := nSub*sx, nSub*sy, nSub*sz
	j := 0
	for jz := 0; jz < nSub; jz++ {
		for jy := 0; jy < nSub; jy++ {
			for jx := 0; jx < nSub; jx++ {
				ix, iy, iz := jx+ix0, jy+iy0, jz+iz0
				i := ix + iy*nFile + iz*nFile*nFile
				for k := 0; k < 3; k++ { x[i][k] = subBuf[k][j] }
				j++
			}
		}
	}
}

// bounds returns the minimum and maximum of an array of vectors.
func bounds(vec [][3]float32) (min, max [3]float32) {
	min, max = vec[0], vec[0]
	for i := range vec {
		for k := 0; k < 3; k++ {
			if vec[i][k] < min[k] { min[k] = vec[i][k] }
			if vec[i][k] > max[k] { max[k] = vec[i][k] }
		}
	}
	return min, max
}

func boolToByte(b bool) byte {
	if b { return 1 }
	return 0
}

func byteToBool(b byte) bool {
	return b != 0
}
