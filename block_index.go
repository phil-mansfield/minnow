package minnow

import (
	"fmt"
)

type blockIndex struct {
	startBlock int
	offsets []int64
}

func newBlockIndex(startBlock int) *blockIndex {
	return &blockIndex{ startBlock, []int64{}, }
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
	if b < idx.startBlock || b >= idx.startBlock + len(idx.offsets) {
		panic(fmt.Sprintf("Group contains blocks in range [%d, %d), but block" +
			" %d was requested.", idx.startBlock,
			idx.startBlock + len(idx.offsets), b))
	}

	return idx.offsets[b - idx.startBlock]
}
