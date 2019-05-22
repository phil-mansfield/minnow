from __future__ import division, print_function

import numpy as np
import minnow
import struct
import bit
import time

def create_int_record(fname, text, xs):
    f = minnow.create(fname)

    f.header(struct.pack("<qq", 0xdeadbeef, len(xs)))
    f.header(text)
    for i in range(len(xs)):
        f.fixed_size_group(np.int64, len(xs[i]))
        f.data(xs[i])
    f.header(np.array(map(len, xs), dtype=np.int64))

    f.close()

def create_group_record(fname, ix, fx, text):
    f = minnow.create(fname)

    ni, nf = len(ix)/4, len(fx)/2
    f.header(struct.pack("<qq", 4, ni))
    f.fixed_size_group(np.int32, ni)
    for i in range(4):
        f.data(ix[i*ni: (i+1)*ni])

    f.header(struct.pack("<qq", 2, nf))
    f.fixed_size_group(np.float64, nf)
    for i in range(2):
        f.data(fx[i*nf: (i+1)*nf])

    f.header(text)

    f.close()

def read_int_record(fname):
    f = minnow.open(fname)

    magic, blocks = f.header(0, "qq")
    text = f.header(1, "s")
    lengths = f.header(2, np.int64)

    xs = [f.data(i) for i in range(blocks)] 

    return text, xs

def read_group_record(fname):
    f = minnow.open(fname)

    bi, ni = f.header(0, "qq")
    bf, nf = f.header(1, "qq")
    text = f.header(2, "s")

    xi = np.zeros(ni*bi, dtype=np.int64)
    xf = np.zeros(nf*bf, dtype=np.float32)
    for i in range(bi):
        xi[i*ni: (i+1)*ni] = f.data(i)
    for i in range(bf):
        xf[i*nf: (i+1)*nf] = f.data(i + bi)

    return xi, xf, text

def test_int_record():
    fname = "../test_files/int_record.test"
    xs = [np.array([1, 2, 3, 4], dtype=np.int64),
          np.array([5], dtype=np.int64),
          np.array([6, 7, 8, 9], dtype=np.int64),
          np.array([10, 11, 12], dtype=np.int64)]
    text = b"I am a cat and I like to meow."

    create_int_record(fname, text, xs)
    rd_text, rd_xs = read_int_record(fname)

    assert(rd_text == text)
    for i in range(len(xs)):
        assert(np.all(xs[i] == rd_xs[i]))

def test_group_record():
    fname = "../test_files/group_files.test"
    ix = np.arange(20, dtype=np.int32)
    fx = np.array(np.arange(10) / 10.0, dtype=np.float64)
    text = b"I'm a caaaat"

    create_group_record(fname, ix, fx, text)
    rd_ix, rd_fx, rd_text = read_group_record(fname)
    
    assert(text == rd_text)
    assert(np.all(rd_ix == ix))
    assert(np.all(np.abs(fx - rd_fx) < 1e-6))

def test_bit_array():
    bits = np.arange(7, 64, dtype=np.int)

    x = np.arange(100, dtype=np.int) 

    for b in bits:
        arr = bit.array(b, x)
        y = bit.from_array(arr, b, len(x))
        assert(np.all(x == y))

def bench_bit_array():
    x = np.arange(100000, dtype=np.uint64) % 100
    N = 1000

    for bits in [8, 11, 16, 23, 32, 45, 64]:
        t0 = time.time()
        for _ in range(N):
            bit.array(bits, x)
        t1 = time.time()
        dt = (t1 - t0) / N
        print("%d bits: %g MB/s" % (bits,  (8*len(x)/ dt) / 1e6))


def read_bit_int_record(fname):
    f = minnow.open(fname)
    
    x2_len = f.header(0, np.int64)
    x1 = f.data(0)
    x2 = [None]*x2_len
    for i in range(x2_len): x2[i] = f.data(1 + i)
    x3 = f.data(x2_len + 1)

    f.close()

    return x1, x2, x3

def test_bit_int_record():
    fname = "../test_files/bit_int_record.test"
    x1 = np.array([100, 101, 102, 104], dtype=int)
    x2 = [np.array([1024, 1024, 1024]), np.array([0, 1023, 500])]
    x3 = np.array([-1000000, -500000])

    rd_x1, rd_x2, rd_x3 = read_bit_int_record(fname)
    
    assert(np.all(x1 == rd_x1))
    assert(np.all(rd_x2[0] == x2[0]))
    assert(np.all(rd_x2[1] == x2[1]))
    assert(np.all(rd_x3 == x3))

if __name__ == "__main__":
    test_int_record()
    test_group_record()
    test_bit_array()
    test_bit_int_record()

    #bench_bit_array()
