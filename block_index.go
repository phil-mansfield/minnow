package minnow

import (
	"fmt"
)

type blockIndex struct {
	startBlock int64
	offsets []int64
}

func newBlockIndex(startBlock int) *blockIndex {
	return &blockIndex{ int64(startBlock), []int64{}, }
}

func (idx *blockIndex) addBlock(size int64) {
	if len(idx.offsets) == 0 {
		idx.offsets = []int64{size}
		return
	}
	prevOffset := idx.offsets[len(idx.offsets) - 1]
	idx.offsets = append(idx.offsets, size + prevOffset)
}

func (idx *blockIndex) blockOffset(b int) int64 {
	b64 := int64(b)
	if b64 < idx.startBlock || b64 >= idx.startBlock + int64(len(idx.offsets)) {
		panic(fmt.Sprintf("Group contains blocks in range [%d, %d), but block" +
			" %d was requested.", idx.startBlock,
			idx.startBlock + int64(len(idx.offsets)), b64))
	}

	return idx.offsets[b64 - idx.startBlock]
}

func (idx *blockIndex) blocks() int64 {
	return int64(len(idx.offsets))
}
