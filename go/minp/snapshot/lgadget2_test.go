package snapshot

import (
	"testing"
)

func newTestMockSnapshot() Snapshot {
	nSide := 10
	nSide3 := nSide*nSide*nSide

	hd := &Header{
		Z: 1, Scale: 0.5,
		OmegaM: 0.3, OmegaL: 0.7, H100: 0.7,
		L: 10, Epsilon: 0.1,
		NSide: int64(nSide), NTotal: int64(nSide3),
		UniformMp: 8.3255525e+10,
	}

	x := [][][3]float32{ make([][3]float32, nSide3) }
	v := [][][3]float32{ make([][3]float32, nSide3) }
	id := [][]int64{ make([]int64, nSide3) }

	i := 0
	for iz := 0; iz < int(hd.NSide); iz++ {
		for iy := 0; iy < int(hd.NSide); iy++ {
			for ix := 0; ix < int(hd.NSide); ix++ {
				x[0][i] = [3]float32{ float32(ix), float32(iy), float32(iz) }
				v[0][i] = [3]float32{ -float32(ix), float32(iy), -float32(iz) }
				id[0][i] = int64(i+1)
				i++
			}
		}
	}

	return NewMockSnapshot(hd, x, v, id)
}

func writeTestLGadget2Snapshot() *lGadget2Header {
	nSide := 10
	nSide3 := nSide*nSide*nSide

	lgHd := &lGadget2Header{
		NPartTotal:      [6]uint32{0, uint32(nSide3), 0, 0, 0, 0},
		NPart: [6]uint32{1, 1, 1, 1, 1, 1},
		Mass: [6]float64{0, 1e10, 0, 0, 0, 0},
		Time: 0.5, Redshift: 1,
		BoxSize: 10, Omega0: 0.3, OmegaLambda: 0.7, HubbleParam: 0.7,
	}
	
	snap := newTestMockSnapshot()
	WriteLGadget2("test_lgadget2_data", "test.%03d", snap, lgHd)

	return lgHd
}

func TestReadWriteLGadget2(t *testing.T) {
	writeTestLGadget2Snapshot()
	snap := newTestMockSnapshot()

	lsnap, err := LGadget2("test_lgadget2_data")
	if err != nil { panic(err.Error()) }

	x, _ := snap.ReadX(0)
	lx, err := lsnap.ReadX(0)
	if err != nil { panic(err.Error()) }
	for i := range lx {
		if !vecEq(lx[i], x[i], 1e-4) {
			t.Errorf("snap.X[%d] = %g, not %g", i, lx[i], x[i])
		}
	}

	v, _ := snap.ReadV(0)
	lv, err := lsnap.ReadV(0)
	if err != nil { panic(err.Error()) }
	for i := range lv {
		if !vecEq(lv[i], v[i], 1e-4) {
			t.Errorf("snap.V[%d] = %g, not %g", i, lv[i], v[i])
		}
	}

	id, _ := snap.ReadID(0)
	lid, err := lsnap.ReadID(0)
	if err != nil { panic(err.Error()) }
	for i := range lid {
		if id[i] != lid[i] {
			t.Errorf("snap.ID[%d] = %d, not %d", i, id[i], lid[i])
		}
	}
}

func vecEq(x, y [3]float32, eps float32) bool {
	return floatEq(x[0], y[0], eps) &&
		floatEq(x[1], y[1], eps) &&
		floatEq(x[2], y[2], eps)
}
