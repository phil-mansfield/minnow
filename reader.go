package minnow

//////////////////
// MinnowReader //
//////////////////

type MinnowReader struct {
	
}

func Open(fname string) *MinnowReader {
	panic("NYI")
}

func (rd *MinnowReader) FileType() FileType {
	panic("NYI")
}

func (rd *MinnowReader) Header(i int, out interface{}) {
	panic("NYI")
}

func (rd *MinnowReader) HeaderSize(i int) int {
	panic("NYI")
}

func (rd *MinnowReader) Data(i int, out interface{}) {
	panic("NYI")
}

func (rd *MinnowReader) Close() {
	panic("NYI")
}
