package snapshot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"runtime"

	"unsafe"
)

type lGadget2Snapshot struct {
	hd Header
	context LGadget2Context
	filenames []string

	xBuf, vBuf [][3]float32
	mpBuf []float32
	idBuf []int64	
}

type LGadget2Context struct {
	NPartNum int
	Order binary.ByteOrder
}

var defaultLGadget2Context = LGadget2Context{
	Order: binary.LittleEndian,
	NPartNum: 2,
}

// LGadget2Snapshot returns a snapshot for the LGadget-2 files in a given
// directory. Additional information may be optionall offered in the form of
// an LGadget2Context instance.
func LGadget2(
	dir string, context ...LGadget2Context,
) (Snapshot, error) {
	snap := &lGadget2Snapshot{ }
	var err error

	snap.context = defaultLGadget2Context
	if len(context) > 0 { snap.context = context[0] }

	snap.filenames, err = getFilenames(dir)
	if err != nil { return nil, err } 
	if len(snap.filenames) == 0 {
		return nil, fmt.Errorf("No files in director %s", dir)
	}

	hd, err := readLGadget2Header(snap.filenames[0], snap.context.Order)
	if err != nil { return nil, err }
	snap.hd = *hd.convert(snap.context.NPartNum)

	return snap, nil
}

// getFilenames returns the names of all the files in a directory.
func getFilenames(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil { return nil, err }

	out := make([]string, len(files))
	for i := range files {
		out[i] = path.Join(dir, files[i].Name())
	}

	return out, nil
}

func readLGadget2Header(
	file string, order binary.ByteOrder,
) (*lGadget2Header, error) {

	out := &lGadget2Header{ }

	f, err := os.Open(file)
	if err != nil { return nil, err }
	defer f.Close()

	_ = readInt32(f, order)
	err = binary.Read(f, binary.LittleEndian, out)
	return out, err
}

func (gh *lGadget2Header) convert(nPartNum int) *Header {
	// Assumes the catalog has already been checked for corruption.
	
	hd := &Header{ }
	
	hd.Z = gh.Redshift
	hd.Scale = 1/(1 + hd.Z)
	hd.L = gh.BoxSize
	hd.OmegaM = gh.Omega0
	hd.OmegaL = gh.OmegaLambda
	hd.H100 = gh.HubbleParam

	hd.NTotal = lgadgetParticleNum(gh.NPartTotal, gh, nPartNum)
	hd.NSide = intCubeRoot(hd.NTotal)

	hd.calcUniformMass()

	return hd
}

func lgadgetParticleNum(
	npart [6]uint32, gh *lGadget2Header, nPartNum int,
) int64 {
	if nPartNum == 2 {
		if npart[0] > 100 * 1000 {
			panic(
				"Simulation contains too many particles. This is probably " +
				"because GadgetNpartNum is set to 2 when it " +
				"should be set to 1.",
			)
		}
		return int64(npart[1]) + int64(uint32(npart[0])) << 32
	} else {
		return int64(npart[0])
	}
}

func intCubeRoot(n int64) int64 {
	c := math.Pow(float64(n), 1.0/3)
	hi, lo := math.Ceil(c), math.Floor(c)
	if hi - c < c - lo {
		return int64(hi)
	} else {
		return int64(lo)
	}
}

// readInt32 returns single 32-bit interger from the given file using the
// given endianness.
func readInt32(r io.Reader, order binary.ByteOrder) int32 {
	var n int32
	if err := binary.Read(r, order, &n); err != nil {
		panic(err.Error())
	}
	return n
}

func writeInt32(w io.Writer, order binary.ByteOrder, x int32) {
	if err := binary.Write(w, order, x); err != nil{
		panic(err.Error())
	}
}


func (snap *lGadget2Snapshot) Files() int {
	return len(snap.filenames)
}

func (snap *lGadget2Snapshot) Header() *Header {
	return &snap.hd
}

func (snap *lGadget2Snapshot) RawHeader(idx int) []byte {
	f, err := os.Open(snap.filenames[idx])
	if err != nil { panic(err.Error()) }
	defer f.Close()

	order := snap.context.Order

	// Read header
	_ = readInt32(f, order)
	buf := make([]byte, unsafe.Sizeof(lGadget2Header{}))
	err = binary.Read(f, order, buf)

	return buf
}

func (snap *lGadget2Snapshot) UpdateHeader(hd *Header) {
	snap.hd = *hd
}

func (snap *lGadget2Snapshot) ReadX(idx int) ([][3]float32, error) {
	f, err := os.Open(snap.filenames[idx])
	if err != nil { return nil, err }
	defer f.Close()

	gh := &lGadget2Header{}
	order := snap.context.Order

	// Read header
	_ = readInt32(f, order)
	binary.Read(f, order, gh)
	_ = readInt32(f, order)

	count := lgadgetParticleNum(gh.NPart, gh, snap.context.NPartNum)
	snap.xBuf = expandVectors(snap.xBuf[:0], int(count))

	// Read position data
	_ = readInt32(f, order)
	readVecAsByte(f, order, snap.xBuf)
	_ = readInt32(f, order)

	L := float32(gh.BoxSize)
	for i := range snap.xBuf {
		for j := 0; j < 3; j++ {
			
			x := snap.xBuf[i][j]

			if x < 0 {
				x += L
			} else if x >= L {
				x -= L
			}

			if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) ||
				x < 0 || x >= L {
				return nil, fmt.Errorf(
					"Corruption detected in the file %s.", snap.filenames[i],
				)
			}

			snap.xBuf[i][j] = x
		}
	}

	return snap.xBuf, nil
}

func (snap *lGadget2Snapshot) UniformMass() bool { return true }

func (snap *lGadget2Snapshot) ReadV(idx int) ([][3]float32, error) {
	f, err := os.Open(snap.filenames[idx])
	if err != nil { return nil, err }
	defer f.Close()

	gh := &lGadget2Header{}
	order := snap.context.Order

	// Read header
	_ = readInt32(f, order)
	binary.Read(f, order, gh)
	_ = readInt32(f, order)

	count := lgadgetParticleNum(gh.NPart, gh, snap.context.NPartNum)
	snap.vBuf = expandVectors(snap.vBuf[:0], int(count))

	// Skip to start of velocity block
	_, err = f.Seek(int64(8 + count*12), 1)
	if err != nil { return nil, err }

	// Read data
	_ = readInt32(f, order)
	readVecAsByte(f, order, snap.vBuf)
	_ = readInt32(f, order)

	rootA := float32(math.Sqrt(float64(gh.Time)))

	// Update units and check data.
	for i := range snap.vBuf {
		for j := 0; j < 3; j++ {
			v := snap.vBuf[i][j] * rootA
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return nil, fmt.Errorf(
					"Corruption detected in the file %s.", snap.filenames[i],
				)
			}
			snap.vBuf[i][j] = v
		}
	}

	return snap.vBuf, nil
}

func (snap *lGadget2Snapshot) ReadID(idx int) ([]int64, error) {
	f, err := os.Open(snap.filenames[idx])
	if err != nil { return nil, err }
	defer f.Close()

	gh := &lGadget2Header{}
	order := snap.context.Order

	// Read header
	_ = readInt32(f, order)
	binary.Read(f, order, gh)
	_ = readInt32(f, order)

	count := lgadgetParticleNum(gh.NPart, gh, snap.context.NPartNum)
	snap.idBuf = expandInts(snap.idBuf[:0], int(count))

	// Skip to start of ID block
	_, err = f.Seek(int64(16 + count*24), 1)
	if err != nil { return nil, err }

	// Read IDs.
	_ = readInt32(f, order)
	readInt64AsByte(f, order, snap.idBuf)
	_ = readInt32(f, order)

	return snap.idBuf, nil
}

func (snap *lGadget2Snapshot) ReadMp(idx int) ([]float32, error) {
	f, err := os.Open(snap.filenames[idx])
	if err != nil { return nil, err }
	defer f.Close()

	gh := &lGadget2Header{}
	order := snap.context.Order

	// Read header
	_ = readInt32(f, order)
	binary.Read(f, order, gh)
	_ = readInt32(f, order)

	count := lgadgetParticleNum(gh.NPart, gh, snap.context.NPartNum)
	snap.mpBuf = expandScalars(snap.mpBuf[:0], int(count))

	for i := range snap.mpBuf {
		snap.mpBuf[i] = float32(snap.hd.UniformMp)
	}

	return snap.mpBuf, nil
}

// gadgetHeader is the formatting for meta-information used by Gadget 2.
type lGadget2Header struct {
	NPart                                     [6]uint32
	Mass                                      [6]float64
	Time, Redshift                            float64
	FlagSfr, FlagFeedback                     int32
	NPartTotal                                [6]uint32
	FlagCooling, NumFiles                     int32
	BoxSize, Omega0, OmegaLambda, HubbleParam float64
	FlagStellarAge, HashTabSize               int32

	Padding [88]byte
}

func expandVectors(vecs [][3]float32, n int) [][3]float32 {
	switch {
	case cap(vecs) >= n:
		return vecs[:n]
	case int(float64(cap(vecs))*1.5) > n:
		return append(vecs[:cap(vecs)],
			make([][3]float32, n-cap(vecs))...)
	default:
		return make([][3]float32, n)
	}
}

func expandScalars(scalars []float32, n int) []float32 {
	switch {
	case cap(scalars) >= n:
		return scalars[:n]
	case int(float64(cap(scalars))*1.5) > n:
		return append(scalars[:cap(scalars)],
			make([]float32, n-cap(scalars))...)
	default:
		return make([]float32, n)
	}
}

func expandInts(ints []int64, n int) []int64 {
	switch {
	case cap(ints) >= n:
		return ints[:n]
	case int(float64(cap(ints))*1.5) > n:
		return append(ints[:cap(ints)], make([]int64, n-cap(ints))...)
	default:
		return make([]int64, n)
	}
}
 
func BytesToLGadget2Header(b []byte) *lGadget2Header {
	if len(b) != int(unsafe.Sizeof(lGadget2Header{})) {
		panic(fmt.Sprintf("length of buffer = %d, but Sizeof(lGadget2Header)"+
			" = %d",len(b), unsafe.Sizeof(lGadget2Header{})))
	}

	hd := &lGadget2Header{ }
	binary.Read(bytes.NewBuffer(b), binary.LittleEndian, hd)
	return hd
}

func WriteLGadget2(
	dir, fnameFmt string, snap Snapshot, hd *lGadget2Header,
) error {

	rootA := float32(math.Sqrt(float64(hd.Time)))

	for i := 0; i < snap.Files(); i++ {
		f, err := os.Create(path.Join(dir, fmt.Sprintf(fnameFmt, i)))
		if err != nil { return err }

		runtime.GC()

		x, err := snap.ReadX(i)
		if err != nil { panic(err.Error()) }
		hd.NPart = [6]uint32{ }
		hd.NPart[1] = uint32(len(x))

		headerSize := int32(unsafe.Sizeof(*hd))
		writeInt32(f, binary.LittleEndian, headerSize)
		err = binary.Write(f, binary.LittleEndian, hd)
		if err != nil { panic(err.Error()) }
		writeInt32(f, binary.LittleEndian, headerSize)

		xSize := int32(12 * len(x))
		writeInt32(f, binary.LittleEndian, xSize)
		err = binary.Write(f, binary.LittleEndian, x)
		if err != nil { panic(err.Error()) }
		writeInt32(f, binary.LittleEndian, xSize)

		runtime.GC()

		v, err := snap.ReadV(i)
		if err != nil { panic(err.Error()) }

		for i := range v {
			for j := 0; j < 3; j++ { v[i][j] /= rootA }
		}

		writeInt32(f, binary.LittleEndian, xSize)
		err = binary.Write(f, binary.LittleEndian, v)
		if err != nil { panic(err.Error()) }
		writeInt32(f, binary.LittleEndian, xSize)

		for i := range v {
			for j := 0; j < 3; j++ { v[i][j] *= rootA }
		}

		runtime.GC()

		id, err := snap.ReadID(i)
		if err != nil { panic(err.Error()) }

		idSize := int32(8*len(id))
		writeInt32(f, binary.LittleEndian, idSize)
		err = binary.Write(f, binary.LittleEndian, id)		
		if err != nil { panic(err.Error()) }
		writeInt32(f, binary.LittleEndian, idSize)

		f.Close()
	}

	return nil
}
