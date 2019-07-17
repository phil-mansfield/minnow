package snapshot

import (
	"fmt"
	"path"

	"github.com/phil-mansfield/minnow/go/minp"
)

type MinP struct {
	fileCells int
	hd *minp.Header
	rawHd []byte
	dir, fileFmt string

	vecBuffer [][3]float32
	idBuffer []int64
	mpBuffer []float32
}

func NewMinP(dir, fileFmt string, fileCells, subCells int) (Snapshot, error) {
	m := &MinP{ dir: dir, fileFmt: fileFmt }
	f0 := minp.Open(m.fileName("x", 0))
	defer f0.Close()

	m.fileCells = f0.FileCells
	m.rawHd = f0.RawHeader
	hd := f0.Header
	m.hd = &hd

	nFile := int(m.hd.NSide) / m.fileCells
	nFile3 := nFile*nFile*nFile

	m.vecBuffer = make([][3]float32, nFile3)
	m.idBuffer, m.mpBuffer = make([]int64, nFile3), make([]float32, nFile3)

	return m, nil
}

func (m *MinP) Files() int {
	return m.fileCells*m.fileCells*m.fileCells
}

func (m *MinP) Header() *minp.Header {
	return m.hd
}

func (m *MinP) RawHeader(i int) []byte {
	return m.rawHd
}

func (m *MinP) UpdateHeader(hd *minp.Header) {
	m.hd = hd
}

func (m *MinP) UniformMass() bool {
	return true
}

func (m *MinP) ReadX(i int) ([][3]float32, error) {
	f := minp.Open(m.fileName("x", i))
	f.Vectors(m.vecBuffer)
	f.Close()
	return m.vecBuffer, nil
}

func (m *MinP) ReadV(i int) ([][3]float32, error) {
	f := minp.Open(m.fileName("v", i))
	f.Vectors(m.vecBuffer)
	f.Close()
	return m.vecBuffer, nil
}

func (m *MinP) ReadID(i int) ([]int64, error) {
	f := minp.Open(m.fileName("x", i))
	f.IDs(m.idBuffer)
	f.Close()
	return m.idBuffer, nil
}

func (m *MinP) ReadMp(i int) ([]float32, error) {
	for i := range m.mpBuffer { m.mpBuffer[i] = float32(m.hd.UniformMp) }
	return m.mpBuffer, nil
}

func (m *MinP) fileName(v string, i int) string {
	return path.Join(m.dir, fmt.Sprintf(m.fileFmt, v, i))
}

func WriteMinP(
	dir, fnameFmt string,
	fileCells, subCells int,
	dx, dv float64,
	snap Snapshot,
) {
	snap = NewGrid(snap, fileCells)

	for i := 0; i < snap.Files(); i++ {
		c := minp.Cell{ int64(i), int64(fileCells), int64(subCells) }
        f := minp.Create(path.Join(dir, fmt.Sprintf(fnameFmt, "x", i)))

        f.Header(snap.Header(), snap.RawHeader(i), c, dx, true)

        x, err := snap.ReadX(i)
        if err != nil { panic(err.Error())  }
        f.Vectors(x)

		f.Close()
	}

	for i := 0; i < snap.Files(); i++ {
		c := minp.Cell{ int64(i), int64(fileCells), int64(subCells) }
        f := minp.Create(path.Join(dir, fmt.Sprintf(fnameFmt, "v", i)))

        f.Header(snap.Header(), snap.RawHeader(i), c, dv, false)

        v, err := snap.ReadX(i)
        if err != nil { panic(err.Error())  }
        f.Vectors(v)

		f.Close()		
	}
}
