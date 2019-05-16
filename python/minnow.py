from __future__ import division, print_function

import numpy as np
import numpy.random as random
import struct
import sys
import copy
import abc

MAGIC = 0xacedad
VERSION = 1

int64_group = 0
int32_group = 1
int16_group = 2
int8_group = 3
uint64_group = 4
uint32_group = 5
uint16_group = 6
uint8_group = 7
float64_group = 8
float32_group = 9 

_py_open = open

def open(fname): return MinnowReader(fname)

class MinnowReader(object):
    def __init__(self, fname):
        self.f = _py_open(fname, "rb")
        f = self.f
        min_hd = struct.unpack("<qqqqqq", f.read(6*8))
        magic, version, groups, headers, blocks, tail_start = min_hd

        assert(MAGIC == magic)
        assert(VERSION == version)

        self.groups, self.headers, self.blocks = groups, headers, blocks
        self.f.seek(tail_start)

        dtype = np.dtype(np.int64).newbyteorder("<")
        self.header_offsets = np.frombuffer(f.read(8*headers), dtype=dtype)
        self.header_sizes = np.frombuffer(f.read(8*headers), dtype=dtype)
        self.group_offsets = np.frombuffer(f.read(8*groups), dtype=dtype)

        self.group_types = np.frombuffer(f.read(8*groups), dtype=dtype)
        group_blocks = np.frombuffer(f.read(8*groups), dtype=dtype)

        readers = [None]*groups
        for i in range(groups):
            readers[i] = _group_from_tail(f, self.group_types[i])

        self.block_index = np.zeros(blocks, dtype=np.int64)
        i0 = 0
        for i in range(groups):
            idx = np.ones(group_blocks[i], dtype=np.int64)*i
            self.block_index[i0: i0+group_blocks[i]] = idx
            i0 += group_blocks[i]

    def header(self, i):
        self.f.seek(self.header_offsets[i], 0)
        return self.f.read(self.header_sizes[i])
        
    def blocks(self):
        return self.blocks

    def data(self):
        i = self.block_index[b]
        self.f.seek(self.group_offsets, 0)
        self.f.seek(self.readers[i].block_offset(b), 1)
        return self.readers[i].read_data(self.f)
        
        

class _Group:
    __metaclass__ = abc.ABCMeta
    @abc.abstractmethod
    def group_type(self): pass
    @abc.abstractmethod
    def write_data(self, f, x): pass
    @abc.abstractmethod
    def write_tail(self, f, x): pass
    @abc.abstractmethod
    def block_offset(self, b): pass
    @abc.abstractmethod
    def read_data(self, f, out): pass

def _group_from_tail(f, gt):
    if gt >= int64_group and gt <= float64_group:
        return new_fixed_size_group_from_tail(f, gt)
    assert(0)

_fixed_size_bytes = [8, 4, 2, 1, 8, 4, 2, 1, 8, 4]
_fixed_size_dtypes = [
    np.dtype(np.int64).newbyteorder("<"),
    np.dtype(np.int32).newbyteorder("<"),
    np.dtype(np.int16).newbyteorder("<"),
    np.dtype(np.int8).newbyteorder("<"),
    np.dtype(np.uint64).newbyteorder("<"),
    np.dtype(np.uint32).newbyteorder("<"),
    np.dtype(np.uint16).newbyteorder("<"),
    np.dtype(np.uint8).newbyteorder("<"),
    np.dtype(np.float64).newbyteorder("<"),
    np.dtype(np.float32).newbyteorder("<")
]

class _BlockIndex(object):
    def __init__(self, start_block):
        self.start_block = start_block
        self.offsets = []

    def add_block(self, size):
        if len(self.offsets) == 0:
            self.offsets = [size]
            return

    def block_offset(self, b):
        if b < self.start_block or b >= self.start_block + len(self.offsets):
            print(("Group contains blocks in range [%d, %d), but block %d" + 
                   "was requested") % (self.start_block,
                                       self.start_block + len(self.offsets), b))
            assert(0)
        if b == self.start_block: return 0
        return self.offsets[b - self.start_block - 1]
    
    def blocks(self):
        return len(self.offsets)

class _FixedSizeGroup(_Group, _BlockIndex):
    def __init__(self, start_block, N, gt):
        _BlockIndex.__init__(self, start_block)
        self.N = N
        self.gt = gt
        self.type_size = _fixed_size_bytes[gt]

    def group_type(self): return self.gt

    def write_data(self, f, x):
        x = np.asarray(x, _fixed_size_dtype[self.gt])
        f.write(x.tobytes())
        self.add_block(self.type_size*self.N)

    def write_tail(self, f, x):
        f.write(struct.pack("<qqq", self.N, self.start_block, self.blocks()))

    def read_data(self, f):
        dtype = _fixed_size_dtype[self.gt]
        return np.frombuffer(fp.read(self.N*self.type_size), dtype=dtype)

    def block_offset(self, b):
        return _Block_Index.block_offset(self, b)


def new_fixed_size_group_from_tail(f, gt):
    N, start_block, blocks = struct.unpack("<qqq", f.read(24))
    g = _FixedSizeGroup(start_block, N, gt)
    for i in range(blocks):
        g.add_block(g.type_size)
    return g

if __name__ == "__main__":
    open("../test_files/int_record.test")
    
