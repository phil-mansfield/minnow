package minp

import (
	"testing"
)

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
