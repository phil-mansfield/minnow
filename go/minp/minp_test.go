package minp

import (
	"testing"
)

func TestVecReaderWriter(t *testing.T) {
	dx := 0.1
	tests := []struct{
		nSide, fileCells, subCells int
	} {
		{1, 1, 1},
		{2, 1, 1},
		{8, 1, 1},
		{10, 1, 1},
		{2, 1, 2},
		{10, 1, 2},
		{10, 1, 5},
	}

	hd := &Header{
		Z: 1.0,
		Scale: 0.5,
		OmegaM: 0.27,
		OmegaL: 0.73,
		H100: 0.7,
		Epsilon: 2.0,
		UniformMp: 1e10,
		L: 100.0,
	}

	rawHd := make([]byte, 130)
	for i := range rawHd { rawHd[i] = byte(i) }

	for i := range tests {
		hd.NSide = int64(tests[i].nSide)
		hd.NTotal = hd.NSide*hd.NSide*hd.NSide
		nFile := int(hd.NSide) / tests[i].fileCells

		c := Cell{ 0, int64(tests[i].fileCells), int64(tests[i].subCells) }

		for _, periodic := range []bool{ false, true } {
			//vec := makeVectors([3]float32{20.5, 30.6, 40.7}, 100, nFile)
			vec := makeVectors([3]float32{0, 0, 0}, 100, nFile)
			wr := Create("test_files/test.minp")
			wr.Header(hd, rawHd, c, dx, periodic)
			wr.Vectors(vec)
			wr.Close()

			out := make([][3]float32, len(vec))
			rd := Open("test_files/test.minp")
			rd.Vectors(out)
			
			if !vectorsEq(vec, out, float32(dx)) {
				t.Errorf("%d) vector read failed for nSide = %d, " + 
					"subCells = %d, periodic = %v",
					i, tests[i].nSide, tests[i].subCells, periodic)
			}
			if *hd != rd.Header {
				t.Errorf("%d) Expected header %v, got %v.", *hd, rd.Header)
			}
			if !bytesEq(rawHd, rd.RawHeader) {
				t.Errorf("%d) Expected raw header %v, got %v.",
					rawHd, rd.RawHeader)
			}
			if rd.FileIndex != 0 || 
				rd.FileCells != int(tests[i].fileCells) ||
				rd.Dx != dx || rd.Periodic != periodic {
				t.Errorf("%d) Incorrect header read.")
			}
		}
	}
}

func TestIDs(t *testing.T) {
	nSide, fileCells, subCells := 10, 5, 2
	nFile := nSide / fileCells
	indices := []int{ 0, 3 + 5*2 + 25*1}
	ids := [][]int64{
		{   0,   1,  10,  11, 100, 101, 110, 111 },
		{ 246, 247, 256, 257, 346, 347, 356, 357 },
	}

	for i := range indices {
		wr := Create("test_files/test.minp")
		hd := &Header{ NSide: 10, L: 100 }
		c := Cell{ int64(indices[i]), int64(fileCells), int64(subCells) }
		wr.Header(hd, []byte{}, c, 1.0, true)
		wr.Vectors(make([][3]float32, nFile*nFile*nFile))
		wr.Close()

		rd := Open("test_files/test.minp")
		out := make([]int64, nFile*nFile*nFile)
		rd.IDs(out)

		if !int64sEq(out, ids[i]) {
			t.Errorf("%d) Expected IDs = %d, got %d", i, ids[i], out)
		}
	}
}

func int64sEq(x, y []int64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func bytesEq(b1, b2 []byte) bool {
	if len(b1) != len(b2) { return false }
	for i := range b1 {
		if b1[i] != b2[i] { return false }
	}
	return true
}

func vectorsEq(v1, v2 [][3]float32, dx float32) bool {
	if len(v1) != len(v2) { return false }
	for i := range v1 {
		for k := 0; k < 3; k++ {
			if v1[i][k] + dx < v2[i][k] || v1[i][k] - dx > v2[i][k] {
				return false
			}
		}
	}
	return true
}

func makeVectors(offset [3]float32, L float32, nFile int) [][3]float32 {
	nFile3 := nFile*nFile*nFile
	vec := make([][3]float32, nFile3)
	dx := L / float32(nFile)

	for i := 0; i < nFile3; i++ {
		ix := i % nFile
		iy := (i / nFile) % nFile
		iz := i / (nFile * nFile)

		vec[i] = [3]float32{
			offset[0] + float32(ix)*dx,
			offset[1] + float32(iy)*dx,
			offset[2] + float32(iz)*dx,
		}

		for k := 0; k < 3; k++ {
			if vec[i][k] > L { vec[i][k] -= L }
		}
	}
	return vec
}

func TestGetSubCell(t *testing.T) {
	nSub, subCells := 2, 3
	nFile := nSub*subCells

	nSub3 := nSub*nSub*nSub
	subCells3 := subCells*subCells*subCells	
	nFile3 := nFile*nFile*nFile

	inSubBuf := [3][]float32{
		make([]float32, nSub3), make([]float32, nSub3), make([]float32, nSub3),
	}
	outSubBuf := [3][]float32{
		make([]float32, nSub3), make([]float32, nSub3), make([]float32, nSub3),
	}

	for i := 0; i < nSub3; i++ {
		inSubBuf[0][i] = float32(i)
		inSubBuf[1][i] = float32(i*10)
		inSubBuf[2][i] = float32(i*100)
	}

	vec := make([][3]float32, nFile3)

	for sc := 0; sc < subCells3; sc++ {
		for k := 0; k < 3; k++ {
			for i := range inSubBuf[k] { inSubBuf[k][i] += 1 }
		}

		setSubCell(vec, inSubBuf, sc, subCells, nSub)
		getSubCell(vec, outSubBuf, sc, subCells, nSub)
		
		if !subBufEq(inSubBuf, outSubBuf) {
			t.Fatalf("At sc = %d\nin = %.0f\nout = %.0f",
				sc, inSubBuf, outSubBuf)
		}
	}
}

func subBufEq(buf1, buf2 [3][]float32) bool {
	for k := 0; k < 3; k++ {
		if len(buf1[k]) != len(buf2[k]) { return false }
		for i := range buf1[k] {
			if buf1[k][i] != buf2[k][i] { return false }
		}
	}
	return true
}
