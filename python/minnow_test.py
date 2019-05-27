from __future__ import division, print_function

import numpy as np
import minnow
import minh
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
    f.header(np.array([len(x) for x in xs], dtype=np.int64))

    f.close()

def create_group_record(fname, ix, fx, text):
    f = minnow.create(fname)

    ni, nf = len(ix)//4, len(fx)//2
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

    assert(rd_text == text.decode("ascii"))
    for i in range(len(xs)):
        assert(np.all(xs[i] == rd_xs[i]))

def test_group_record():
    fname = "../test_files/group_files.test"
    ix = np.arange(20, dtype=np.int32)
    fx = np.array(np.arange(10) / 10.0, dtype=np.float64)
    text = b"I'm a caaaat"

    create_group_record(fname, ix, fx, text)
    rd_ix, rd_fx, rd_text = read_group_record(fname)
    
    assert(text.decode("ascii") == rd_text)
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


def write_bit_int_record(fname, x1, x2, x3):
    f = minnow.create(fname)

    f.int_group(len(x1))
    f.data(x1)

    f.header(struct.pack("<q", len(x2)))
    f.int_group(len(x2[0]))
    for i in range(len(x2)): f.data(x2[i])

    f.int_group(len(x3))
    f.data(x3)

    f.close()

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

    write_bit_int_record(fname, x1, x2, x3)
    rd_x1, rd_x2, rd_x3 = read_bit_int_record(fname)
    
    assert(np.all(x1 == rd_x1))
    assert(np.all(rd_x2[0] == x2[0]))
    assert(np.all(rd_x2[1] == x2[1]))
    assert(np.all(rd_x3 == x3))

def create_q_float_record(fname, limit, dx1, dx2, x1, x2):
    f = minnow.create(fname)

    f.header(struct.pack("<ffffqq", dx1, dx2, limit[0], limit[1],
                         len(x1), len(x2)))
    f.float_group(len(x1[0]), limit, dx1)
    for i in range(len(x1)): f.data(x1[i])
    f.float_group(len(x2[0]), limit, dx2)
    for i in range(len(x2)): f.data(x2[i])

    f.close()

def open_q_float_record(fname):
    f = minnow.open(fname)
    
    dx1, dx2, low, high, x1_len, x2_len = f.header(0, "ffffqq")
    x1, x2 = [None]*x1_len, [None]*x2_len
    for i1 in range(x1_len):
        x1[i1] = f.data(i1)
    for i2 in range(x2_len):
        x2[i2] = f.data(i2 + x1_len)

    f.close()

    return x1, x2

def test_q_float_record():
    fname = "../test_files/q_float_record.test"
    limit = (-50, 100)
    dx1, dx2 = 1.0, 10.0
    x1 = [
        np.array([-50, 0, 50, 49]),
        np.array([25, 25, 25, 25])
    ]
    x2 = [
        np.array([-50, 0, 50, 49, 0]),
        np.array([1, 2, 3, 4, 5]),
        np.array([0, 20, 0, 20, 0])
    ]

    create_q_float_record(fname, limit, dx1, dx2, x1, x2)
    rd_x1, rd_x2 = open_q_float_record(fname)

    assert(len(x1) == len(rd_x1))
    for i in range(len(x1)):
        assert(len(x1[i]) == len(rd_x1[i]))
        assert(np.all(eps_eq(x1[i], rd_x1[i], dx1)))

    assert(len(x2) == len(rd_x2))
    for i in range(len(x2)):
        assert(len(x2[i]) == len(rd_x2[i]))
        assert(np.all(eps_eq(x2[i], rd_x2[i], dx2)))

def eps_eq(x, y, eps): return (x + eps > y) & (x - eps < y)

def test_periodic_min():
    pixels = 20
    data = [
        [0, 1, 2, 3],
        [10, 11, 12, 13],
        [18, 19, 0, 1],
        [1, 0, 19, 18],
        [1, 19, 18, 0],
    ]
    mins = [0, 10, 18, 18, 18]

    for i in range(len(data)):
        min = bit.periodic_min(data[i], pixels)
        assert(min == mins[i])

def test_minh_reader_writer():
    fname = "../test_files/reader_writer_minh.test"
    names = ["int64", "float32", "int", "float", "log"]
    text = ("Cats are the best. Don't we love them?!@#$%^&*(),.." +
            "..[]{};':\"|\\/-=_+`~meow meow meow")
    columns = [
        minh.Column(minnow.int64_group),
        minh.Column(minnow.float32_group),
        minh.Column(minnow.int_group),
        minh.Column(minnow.float_group, 0, 100, 200, 1),
        minh.Column(minnow.float_group, 1, 10, 14, 0.01)
    ]

    block1 = [
        np.array([100, 200, 300, 400, 500], dtype=np.int64),
        np.array([150, 250, 350, 450, 550], dtype=np.float32),
        np.array([-30, -35, -25, -10, -20], dtype=np.int64),
        np.array([100, 200, 125, 150, 100], dtype=np.float32),
        np.array([1e10, 1e11, 1e11, 1e14, 3e13], dtype=np.float32)
    ]

    block2 = [
        np.array([125, 225, 325], dtype=np.int64),
        np.array([1750, 2750, 3750], dtype=np.float32),
        np.array([1000, 1000, 1000], dtype=np.int64),
        np.array([100, 100, 100], dtype=np.float32),
        np.array([1e14, 1e14, 1e14], dtype=np.float32)
    ]

    joined_blocks = [np.hstack([block1[i], block2[i]]) for i in range(5)]
    blocks = [block1, block2]

    wr = minh.create(fname)
    wr.header(names, text, columns)
    for block in blocks: wr.block(block)
    wr.close()

    blocks += [joined_blocks]

    rd = minh.open(fname)

    assert(rd.names == names)
    assert(rd.text == text)
    assert(rd.blocks == 2)
    assert(rd.length == 8)
    for i in range(rd.blocks):
        assert(rd.block_lengths[i] == [5, 3][i])
    for i in range(len(columns)):
        assert(column_eq(columns[i], rd.columns[i]))

    for b in range(len(blocks)):
        block = blocks[b]
        if b < 2:
            rd_int64, rd_float32, rd_int, rd_float, rd_log = rd.block(b, names)
        else:
            rd_int64, rd_float32, rd_int, rd_float, rd_log = rd.read(names)

        assert(len(rd_int64) == len(block[0]))
        assert(np.all(rd_int64 == block[0]))

        assert(len(rd_float32) == len(block[1]))
        assert(np.all(eps_eq(rd_float32, block[1], 1e-3)))

        assert(len(rd_int) == len(block[2]))
        assert(np.all(rd_int == block[2]))

        assert(len(rd_float) == len(block[3]))
        assert(np.all(eps_eq(rd_float, block[3], 1)))

        assert(len(rd_log) == len(block[4]))
        assert(np.all(eps_eq(np.log10(rd_log), np.log10(block[4]), 0.01)))

def column_eq(c1, c2):
    return (c1.type == c2.type and c1.log == c2.log and 
            eps_eq(c1.dx, c2.dx, 1e-5) and
            eps_eq(c1.low, c2.low, 1e-5) and
            eps_eq(c1.high, c2.high, 1e-5))

if __name__ == "__main__":
    test_int_record()
    test_group_record()
    test_bit_array()
    test_periodic_min()
    test_bit_int_record()
    test_q_float_record()
    test_minh_reader_writer()

    #bench_bit_array()
