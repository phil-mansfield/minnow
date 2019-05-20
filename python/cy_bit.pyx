import numpy as np
cimport numpy as np
cimport cython

def precision_needed(max):
    return int(np.ceil(np.log2(max + 1)))

def array_bytes(bits, length):
    return int(np.ceil(float(bits * length) / 8))

def array(np.uint64_t bits, np.uint64_t[:] x):
    cdef np.uint8_t[:] b = np.zeros(array_bytes(bits, len(x)), dtype=np.uint8)
    cdef np.uint8_t[:] buf = np.zeros(8, np.uint8)
    cdef np.uint8_t[:] t_buf = np.zeros(9, np.uint8)

    cdef np.uint64_t buf_bytes = np.uint64(bits / 8)
    if buf_bytes * 8 < bits: buf_bytes += 1

    cdef np.uint64_t mask = (~np.uint64(0)) >> (64 - bits)
    
    cdef int i
    cdef int j
    cdef np.uint64_t xi
    cdef np.uint64_t curr_bit
    cdef np.uint64_t start_byte
    cdef np.uint64_t end_byte
    for i in range(len(x)):
        xi = x[i]
        curr_bit = (i*bits) % 8

        for j in range(buf_bytes):
            buf[j] = np.uint8(xi >> 8*j)
            
        t_buf[buf_bytes] = 0
        for j in range(buf_bytes):
            t_buf[j] = buf[j] << curr_bit

        start_byte = (i*bits) / 8
        end_byte = ((i+1)*bits - 1) / 8

        for j in range(end_byte - start_byte + 1):
            b[start_byte + j] |= t_buf[j]

    return b

def slice(np.uint8_t[:] arr, np.uint64_t bits, np.uint64_t length):
    cdef np.uint64_t[:] out = np.zeros(length, dtype=np.uint64)
    cdef np.uint8_t[:] buf = np.zeros(8, np.uint8)
    cdef np.uint8_t[:] t_buf = np.zeros(9, np.uint8)

    cdef np.uint64_t buf_bytes = np.uint64(bits / 8)
    if buf_bytes * 8 < bits: buf_bytes += 1

    cdef np.uint64_t i
    cdef np.uint64_t j
    cdef np.uint64_t start_bit
    cdef np.uint64_t next_start_bit
    cdef np.uint64_t start_byte
    cdef np.uint64_t end_byte
    cdef np.uint64_t t_buf_bytes
    cdef np.uint64_t xi
    cdef np.uint8_t start_mask
    cdef np.uint8_t end_mask
    for i in range(length):
        start_bit = (i*bits) % 8
        next_start_bit = (start_bit + bits) % 8

        start_byte = i*bits / 8
        end_byte = ((i + 1)*bits - 1) / 8
        t_buf_bytes = end_byte - start_byte + 1

        for j in range(t_buf_bytes):
            t_buf[j] = arr[start_byte + j]

        start_mask = ~np.uint8(0) << start_bit
        end_ask = ~np.uint8(0) >> (8 - next_start_bit)
        if next_start_bit == 0: end_mask = ~np.uint8(0)

        t_buf[j] &= start_mask
        t_buf[j] &= end_mask

        for j in range(buf_bytes):
            buf[j] = t_buf[j] >> start_bit
        for j in range(buf_bytes):
            buf[j] |= t_buf[j+1] << (8 - start_bit)

        for i in range(t_buf_bytes): t_buf[i] = 0

        xi = 0
        for j in range(buf_bytes):
            xi |= np.uint64(buf[j]) << (8*j)

    return out
