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

def create(fname): return Writer(fname)

def open(fname): return Reader(fname)

class Writer(object):
    def __init__(self, fname):
        self.f = _py_open(fname, "w+")
        self.headers, self.blocks = 0, 0
        self.writers = []
        self.header_offsets, self.header_sizes = [], []
        self.group_blocks, self.group_offsets = [], []
        self.f.write('\0'*48)

    def header(self, data):
        if type(data) == np.ndarray:
            dtype = np.dtype(data.dtype).newbyteorder("<")
            data = np.asarray(data, dtype).tobytes()
        self.header_offsets.append(self.f.tell())
        self.header_sizes.append(len(data))
        self.f.write(data)

        self.headers += 1
        return self.headers/ - 1

    def fixed_size_group(self, dtype, N):
        group_type = _fixed_size_type_dict[dtype]
        self._new_group(_FixedSizeGroup(self.blocks, N, group_type))

    def _new_group(self, g):
        self.writers.append(g)
        self.group_blocks.append(0)
        self.group_offsets.append(self.f.tell())

    def data(self, data):
        self.writers[-1].write_data(self.f, data)
        self.group_blocks[-1] += 1
        self.blocks += 1
        return self.blocks - 1

    def close(self):
        tail_start = self.f.tell()
        group_types = [g.group_type() for g in self.writers]
        dtype = np.dtype(np.int64).newbyteorder("<")

        self.f.write(np.asarray(self.header_offsets, dtype).tobytes())
        self.f.write(np.asarray(self.header_sizes, dtype).tobytes())
        self.f.write(np.asarray(self.group_offsets, dtype).tobytes())
        self.f.write(np.asarray(group_types, dtype).tobytes())
        self.f.write(np.asarray(self.group_blocks, dtype).tobytes())

        for i in range(len(self.writers)):
            self.writers[i].write_tail(self.f)

        self.f.seek(0, 0)
        self.f.write(struct.pack("<qqqqqq", MAGIC, VERSION, len(self.writers),
                     self.headers, self.blocks, tail_start))

        self.f.close()

class Reader(object):
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

        self.readers = [None]*groups
        for i in range(groups):
            self.readers[i] = _group_from_tail(f, self.group_types[i])

        self.block_index = np.zeros(blocks, dtype=np.int64)
        i0 = 0
        for i in range(groups):
            idx = np.ones(group_blocks[i], dtype=np.int64)*i
            self.block_index[i0: i0+group_blocks[i]] = idx
            i0 += group_blocks[i]

    def header(self, i, data_type):
        self.f.seek(self.header_offsets[i], 0)
        b = self.f.read(self.header_sizes[i])
        if data_type == "s":
            return b.decode("ascii")
        elif type(data_type) == str:
            return struct.unpack("<" + data_type, b)
        elif type(data_type) == type:
            dtype = np.dtype(np.int64).newbyteorder("<")
            return np.frombuffer(b, dtype=dtype)
        
    def blocks(self):
        return self.blocks

    def data(self, b):
        i = self.block_index[b]
        self.f.seek(self.group_offsets[i], 0)
        self.f.seek(self.readers[i].block_offset(b), 1)
        return self.readers[i].read_data(self.f)
        
    def data_type(self, b):
        return self.group_types[self.block_index[b]]

    def close(self):
        self.f.close()

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
        return _new_fixed_size_group_from_tail(f, gt)
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
_fixed_size_type_dict = {
    np.int64: 0, np.int32: 1, np.int16: 2, np.int8: 3,
    np.uint64: 4, np.uint32: 5, np.uint16: 6, np.uint8: 7,
    np.float64: 8, np.float32: 9
}

class _BlockIndex(object):
    def __init__(self, start_block):
        self.start_block = start_block
        self.offsets = []

    def add_block(self, size):
        if len(self.offsets) == 0:
            self.offsets = [size]
            return
        self.offsets.append(size + self.offsets[-1])

    def block_offset(self, b):
        if b < self.start_block or b >= self.start_block + len(self.offsets):
            raise ValueError(
                ("Group contains blocks in range [%d, %d), but block %d " + 
                 "was requested") % (self.start_block,
                                     self.start_block + len(self.offsets), b)
            )
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
        x = np.asarray(x, _fixed_size_dtypes[self.gt])
        f.write(x.tobytes())
        self.add_block(self.type_size*self.N)

    def write_tail(self, f):
        f.write(struct.pack("<qqq", self.N, self.start_block, self.blocks()))

    def read_data(self, f):
        dtype = _fixed_size_dtypes[self.gt]
        return np.frombuffer(f.read(self.N*self.type_size), dtype=dtype)

    def block_offset(self, b):
        return _BlockIndex.block_offset(self, b)


def _new_fixed_size_group_from_tail(f, gt):
    N, start_block, blocks = struct.unpack("<qqq", f.read(24))
    g = _FixedSizeGroup(start_block, N, gt)
    for i in range(blocks):
        g.add_block(g.type_size*g.N)
    return g    
