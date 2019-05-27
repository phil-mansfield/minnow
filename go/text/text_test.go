package text

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func fakeReader(text string, itemSize, blockSize int) *Reader {
	f := bytes.NewReader([]byte(text))
	rd := &Reader{
		f: f, closer: ioutil.NopCloser(f),
		config: DefaultReaderConfig,
	}
	rd.config.MaxItemSize = int64(itemSize)
	rd.config.MaxBlockSize = int64(blockSize)
	return rd
}

func TestLineHeader(t *testing.T) {
	tests := []struct {
		text string
		itemSize int
		lines int
		res string
	} {
		{"a\nb\nc\nd", 100, 2, "a\nb"},
		{"a\nb\nc\nd", 5, 2, "a\nb"},
	}

	for i := range tests {
		rd := fakeReader(tests[i].text, tests[i].itemSize, 1000)
		out := rd.LineHeader(tests[i].lines)

		if out != tests[i].res {
			t.Errorf("Test %d: expected '%s', got '%s'.", i, tests[i].res, out)
		}
	}
}

func TestCommentHeader(t *testing.T) {
	tests := []struct {
		text string
		itemSize int
		res string
	} {
		{"#a\n#b\nc\nd", 100, "#a\n#b"},
		{"#a\n#b\nc\nd", 8, "#a\n#b"},
	}

	for i := range tests {
		rd := fakeReader(tests[i].text, tests[i].itemSize, 1000)
		out := rd.CommentHeader()

		if out != tests[i].res {
			t.Errorf("Test %d: expected '%s', got '%s'.",
				i, tests[i].res, out)
		}
	}
}


func TestReader(t *testing.T) {
	text := []byte(`#123456789012345678
#123456789012345678
1    2     3      5
11  12    13     15
21  22    23     25
31  32    33     35
41  42    43     45
`)
	itemSize := 50
	blockSize := 120

	config := DefaultReaderConfig
	config.MaxItemSize = int64(itemSize)
	config.MaxBlockSize = int64(blockSize)

	names := []string{"1", "2", "3", "4"}
	out1 := []interface{}{ []float32{}, []int64{}, []float32{}, []int64{} }
	out2 := []interface{}{ []float32{}, []int64{}, []float32{}, []int64{} }
	
	exp1 := []interface{} {
		[]float32{5, 15},
		[]int64{1, 11},
		[]float32{2, 12},
		[]int64{3, 13},
	}
	exp2 := []interface{} {
		[]float32{25, 35, 45},
		[]int64{21, 31, 41},
		[]float32{22, 32, 42},
		[]int64{23, 33, 43},
	}

	f := openFromReader(bytes.NewReader(text), config)
	f.SetNames(names)

	if f.Blocks() != 2 {
		t.Errorf("Expected 2 blocks, go %d.", f.Blocks())
	}

	f.Block(0, []string{"4", "1", "2", "3"}, out1)
	f.Block(1, []string{"4", "1", "2", "3"}, out2)

	for i := range exp1 {
		if !genericEq(exp1[i], out1[i], 1e-3) {
			t.Errorf("Expected %v for column %d of block %d, but got %v",
				exp1[i], i, 0, out1[i])
		}
	}
	for i := range exp2 {
		if !genericEq(exp2[i], out2[i], 1e-3) {
			t.Errorf("Expected %v for column %d of block %d, but got %v",
				exp2[i], i, 0, out2[i])
		}
	}
}

func genericEq(x, y interface{}, eps float32) bool {
	switch xs := x.(type) {
	case []int64:
		ys := y.([]int64)
		if len(xs) != len(ys) { return false }
		for i := range xs {
			if xs[i] != ys[i] { return false }
		}
		return true
	case []float32:
		ys := y.([]float32)
		if len(xs) != len(ys) { return false }
		for i := range xs {
			if xs[i] + eps < ys[i] || xs[i] - eps > ys[i] { return false }
		}
		return true
	}
	panic("Bad type")
}

func TestNextBlock(t *testing.T) {
	text := `1234
1234
1234
1234
1234
1234
`
	f := fakeReader(text, 6, 12)
	size := readerSize(f.f)

	for pos := int64(0); pos < size; pos++ {
		expected := pos + 12 - 6
		col := expected % 5
		expected += 5 - col

		if pos + 12 >= 30 { expected = -1 }

		f.f.Seek(pos, 0)
		next := f.nextBlock(size)
		pos2, _ := f.f.Seek(0, 1)

		if next != expected {
			t.Errorf("Expected next block = %d for pos = %d, but got %d",
				expected, pos, next)
		}

		if next != -1 && pos2 != next {
			t.Error("nextBlock did not set position to start of next block.")
		}
	}
}

func TestRockstarHeader(t *testing.T) {
	rockstarLine := "#ID DescID Mvir Vmax Vrms Rvir Rs Np X Y Z VX VY VZ JX JY JZ Spin rs_klypin Mvir_all M200b M200c M500c M2500c Xoff Voff spin_bullock b_to_a c_to_a A[x] A[y] A[z] b_to_a(500c) c_to_a(500c) A[x](500c) A[y](500c) A[z](500c) T/|U| M_pe_Behroozi M_pe_Diemer Halfmass_Radius rvmax PID"
	consistentTreesLine := "#scale(0) id(1) desc_scale(2) desc_id(3) num_prog(4) pid(5) upid(6) desc_pid(7) phantom(8) sam_Mvir(9) Mvir(10) Rvir(11) rs(12) vrms(13) mmp?(14) scale_of_last_MM(15) vmax(16) x(17) y(18) z(19) vx(20) vy(21) vz(22) Jx(23) Jy(24) Jz(25) Spin(26) Breadth_first_ID(27) Depth_first_ID(28) Tree_root_ID(29) Orig_halo_ID(30) Snap_num(31) Next_coprogenitor_depthfirst_ID(32) Last_progenitor_depthfirst_ID(33) Last_mainleaf_depthfirst_ID(34) Tidal_Force(35) Tidal_ID(36) Rs_Klypin(37) Mmvir_all(38) M200b(39) M200c(40) M500c(41) M2500c(42) Xoff(43) Voff(44) Spin_Bullock(45) b_to_a(46) c_to_a(47) A[x](48) A[y](49) A[z](50) b_to_a(500c)(51) c_to_a(500c)(52) A[x](500c)(53) A[y](500c)(54) A[z](500c)(55) T/|U|(56) M_pe_Behroozi(57) M_pe_Diemer(58) Halfmass_Radius(59) Macc(60) Mpeak(61) Vacc(62) Vpeak(63) Halfmass_Scale(64) Acc_Rate_Inst(65) Acc_Rate_100Myr(66) Acc_Rate_1*Tdyn(67) Acc_Rate_2*Tdyn(68) Acc_Rate_Mpeak(69) Acc_Log_Vmax_Inst(70) Acc_Log_Vmax_1*Tdyn(71) Mpeak_Scale(72) Acc_Scale(73) First_Acc_Scale(74) First_Acc_Mvir(75) First_Acc_Vmax(76) Vmax\\@Mpeak(77) Tidal_Force_Tdyn(78) Log_(Vmax/Vmax_max(Tdyn;Tmpeak))(79) Time_to_future_merger(80) Future_merger_MMP_ID(81)"

	rockstarExp := []string{ "ID", "DescID", "Mvir", "Vmax", "Vrms", "Rvir", "Rs", "Np", "X", "Y", "Z", "VX", "VY", "VZ", "JX", "JY", "JZ", "Spin", "rs_klypin", "Mvir_all", "M200b", "M200c", "M500c", "M2500c", "Xoff", "Voff", "spin_bullock", "b_to_a", "c_to_a", "A[x]", "A[y]", "A[z]", "b_to_a(500c)", "c_to_a(500c)", "A[x](500c)", "A[y](500c)", "A[z](500c)", "T/|U|", "M_pe_Behroozi", "M_pe_Diemer", "Halfmass_Radius", "rvmax", "PID" }
	consistentTreesExp := []string{ "scale", "id", "desc_scale", "desc_id", "num_prog", "pid", "upid", "desc_pid", "phantom", "sam_Mvir", "Mvir", "Rvir", "rs", "vrms", "mmp?", "scale_of_last_MM", "vmax", "x", "y", "z", "vx", "vy", "vz", "Jx", "Jy", "Jz", "Spin", "Breadth_first_ID", "Depth_first_ID", "Tree_root_ID", "Orig_halo_ID", "Snap_num", "Next_coprogenitor_depthfirst_ID", "Last_progenitor_depthfirst_ID", "Last_mainleaf_depthfirst_ID", "Tidal_Force", "Tidal_ID", "Rs_Klypin", "Mmvir_all", "M200b", "M200c", "M500c", "M2500c", "Xoff", "Voff", "Spin_Bullock", "b_to_a", "c_to_a", "A[x]", "A[y]", "A[z]", "b_to_a(500c)", "c_to_a(500c)", "A[x](500c)", "A[y](500c)", "A[z](500c)", "T/|U|", "M_pe_Behroozi", "M_pe_Diemer", "Halfmass_Radius", "Macc", "Mpeak", "Vacc", "Vpeak", "Halfmass_Scale", "Acc_Rate_Inst", "Acc_Rate_100Myr", "Acc_Rate_1*Tdyn", "Acc_Rate_2*Tdyn", "Acc_Rate_Mpeak", "Acc_Log_Vmax_Inst", "Acc_Log_Vmax_1*Tdyn", "Mpeak_Scale", "Acc_Scale", "First_Acc_Scale", "First_Acc_Mvir", "First_Acc_Vmax", "Vmax\\@Mpeak", "Tidal_Force_Tdyn", "Log_(Vmax/Vmax_max(Tdyn;Tmpeak))", "Time_to_future_merger", "Future_merger_MMP_ID" }

	rockstarRes := rockstarNames(rockstarLine)
	consistentTreesRes := consistentTreesNames(consistentTreesLine)

	if len(rockstarRes) != len(rockstarExp) {
		t.Errorf("Expected len(rockstar) = %d, but got %d",
			len(rockstarExp), len(rockstarRes))
	}

	if len(consistentTreesRes) != len(consistentTreesExp) {
		t.Errorf("Expected len(consistentTrees) = %d, but got %d",
			len(consistentTreesExp), len(consistentTreesRes))
	}

	for i := range rockstarRes {
		if rockstarRes[i] != rockstarExp[i] {
			t.Errorf("Expected rockstar[%d] = %s, but got %s.",
				rockstarExp[i], i, rockstarRes[i])
		}
	}

	for i := range consistentTreesRes {
		if consistentTreesRes[i] != consistentTreesExp[i] {
			t.Errorf("Expected consistentTrees[%d] = %s, but got %s.",
				consistentTreesExp[i], i, consistentTreesRes[i])
		}
	}
}
