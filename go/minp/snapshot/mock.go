package snapshot

type mockSnapshot struct {
	hd *Header
	x, v [][][3]float32
	mp [][]float32
	id [][]int64
}

func NewMockSnapshot(hd *Header, x, v [][][3]float32, id [][]int64) Snapshot {
	mp := make([][]float32, len(x))
	for i := range mp {
		mp[i] = make([]float32, len(x[0]))
		for j := range mp[i] { mp[i][j] = float32(hd.UniformMp) }
	}

	snap := &mockSnapshot{
		x:x, v:v, id:id, mp:mp, hd:hd,
	}

	return snap
}

func (snap *mockSnapshot) Files() int {
	return len(snap.x)
}
func (snap *mockSnapshot) Header() *Header {
	return snap.hd
}
func (snap *mockSnapshot) RawHeader(i int) []byte {
	return []byte{99}
}
func (snap *mockSnapshot) UpdateHeader(hd *Header) {
	snap.hd = hd
}
func (snap *mockSnapshot) UniformMass() bool {
	return true
}
func (snap *mockSnapshot) ReadX(i int) ([][3]float32, error) {
	return snap.x[i], nil
}
func (snap *mockSnapshot) ReadV(i int) ([][3]float32, error) {
	return snap.v[i], nil
}
func (snap *mockSnapshot) ReadID(i int) ([]int64, error) {
	return snap.id[i], nil
}
func (snap *mockSnapshot) ReadMp(i int) ([]float32, error) {
	return snap.mp[i], nil
}
