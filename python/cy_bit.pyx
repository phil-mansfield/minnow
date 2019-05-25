from __future__ import print_function
import numpy as np
cimport numpy as np
cimport cython

def precision_needed(max):
    return int(np.ceil(np.log2(max + 1)))

def array_bytes(bits, length):
    return int(np.ceil(float(bits * length) / 8))

@cython.boundscheck(False)
@cython.wraparound(False)
@cython.cdivision(True)
def array(np.uint64_t bits, np.uint64_t[:] x):
    cdef np.uint8_t[:] b = np.zeros(array_bytes(bits, len(x)), dtype=np.uint8)
    cdef np.uint8_t[:] buf = np.zeros(8, np.uint8)
    cdef np.uint8_t[:] t_buf = np.zeros(9, np.uint8)

    cdef np.uint64_t buf_bytes = np.uint64(bits / 8)
    if buf_bytes * 8 < bits: buf_bytes += 1

    cdef np.uint64_t mask = (~np.uint64(0)) >> np.uint64(64 - bits)

    cdef Py_ssize_t i, j
    cdef np.uint64_t xi, curr_bit, start_byte, end_byte
    for i in range(len(x)):
        xi = x[i]
        curr_bit = (i*bits) % 8

        for j in range(buf_bytes):
            buf[j] = <np.uint8_t>(xi >> 8*j)
            
        t_buf[buf_bytes] = 0
        for j in range(buf_bytes):
            t_buf[j] = buf[j] << curr_bit

        for j in range(buf_bytes):
            t_buf[j + 1] |= buf[j] >> (8-curr_bit)

        start_byte = (i*bits) / 8
        end_byte = ((i+1)*bits - 1) / 8


        for j in range(end_byte - start_byte + 1):
            b[start_byte + j] |= t_buf[j]

    return np.array(b)

@cython.boundscheck(False)
@cython.wraparound(False)
@cython.cdivision(True)
def from_array(np.uint8_t[:] arr, np.uint64_t bits, np.uint64_t length):
    cdef np.uint64_t[:] out = np.zeros(length, dtype=np.uint64)
    cdef np.uint8_t[:] buf = np.zeros(8, np.uint8)
    cdef np.uint8_t[:] t_buf = np.zeros(9, np.uint8)

    cdef np.uint64_t buf_bytes = np.uint64(bits / 8)
    if buf_bytes * 8 < bits: buf_bytes += 1

    cdef np.uint64_t i, j, xi, start_bit, next_start_bit
    cdef np.uint64_t start_byte, end_byte, t_buf_bytes, eight
    cdef np.uint8_t start_mask, end_mask
    eight = 8 # Don't ask...
    for i in range(length):
        start_bit = (i*bits) % 8
        next_start_bit = (start_bit + bits) % 8

        start_byte = i*bits / 8
        end_byte = ((i + 1)*bits - 1) / 8
        t_buf_bytes = end_byte - start_byte + 1

        for j in range(t_buf_bytes):
            t_buf[j] = arr[start_byte + j]

        start_mask = (0xff << start_bit) & 0xff
        end_mask = (0xff >> (<np.uint8_t>(8 - next_start_bit))) & 0xff
        if next_start_bit == 0: end_mask = 0xff

        t_buf[0] &= start_mask
        t_buf[t_buf_bytes - 1] &= end_mask

        for j in range(buf_bytes):
            buf[j] = t_buf[j] >> start_bit
        for j in range(buf_bytes):
            buf[j] |= t_buf[j+1] << (eight - start_bit)


        for j in range(t_buf_bytes): t_buf[j] = 0

        xi = 0
        for j in range(buf_bytes):
            xi |= (<np.uint64_t>buf[j]) << (<np.uint64_t>(8*j))
        out[i] = xi

    return np.array(out)

@cython.boundscheck(False)
@cython.wraparound(False)
@cython.cdivision(True)
def periodic_min(np.int64_t[:] x, np.int64_t pixels):
    cdef np.int64_t x0 = x[0]
    cdef np.int64_t width = 1
    cdef np.int64_t N = len(x)

    cdef np.int64_t xi, x1, d0, d1
    for i in range(N):
        xi = x[i]
        x1 = x0 + width - 1
        if x1 >= pixels: x1 -= pixels
        
        d0 = periodic_distance(xi, x0, pixels)
        d1 = periodic_distance(xi, x1, pixels)

        if d0 > 0 and d1 < 0: continue 

        if d1 > -d0:
            width += d1
        else:
            x0 += d0
            if x0 < 0: x0 += pixels
            width -= d0

        if width > pixels/2: return 0

    return x0

@cython.boundscheck(False)
@cython.wraparound(False)
@cython.cdivision(True)
cdef np.int64_t periodic_distance(
    np.int64_t x, np.int64_t x0, np.int64_t pixels
):
    cdef np.int64_t d = x - x0
    if d >= 0:
        if d > pixels - d: return d - pixels
    else:
        if d < -(d + pixels): return pixels + d
    return d
