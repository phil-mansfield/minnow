package minnow

import (
	"os"
)

type group interface {
	dataBytes() int64
	tailBytes() int64

	writeData(f *os.File, x interface{})
	writeTail(f *os.File)

	blockOffset(b int) int64

	readData(f *os.File, x interface{})
	readHeader(f *os.File)
}

func newInt64Group(startBlock int, N int) group {
	panic("NYI")
}


type int64Group struct {
	blockIndex
}
