import cy_bit
import numpy as np

def precision_needed(max):
    return cy_bit.precision_needed(np.uint64(max))

def array_bytes(bits, length):
    return cy_bit.array_bytes(np.uint64(bits), np.uint64(length))

def array(bits, x):
    return cy_bit.array(np.uint64(bits), np.asarray(x, dtype=np.uint64))

def from_array(arr, bits, length):
    assert(type(arr) == np.ndarray)
    assert(arr.dtype == np.uint8)
    return cy_bit.from_array(arr, np.uint64(bits), np.uint64(length))

def write_array(f, bits, x):
    if bits == 0: return
    f.write(array(bits, x).tobytes())

def read_array(f, bits, length):
    if bits == 0: return np.zeros(length, dtype=np.uint64)
    buf = np.frombuffer(f.read(array_bytes(bits, length)), dtype=np.uint8)
    # The extra array conversion here is needed to get rid of a read-only bit
    return from_array(np.array(buf, dtype=np.uint8), bits, length)

def periodic_min(x, pixels):
    return cy_bit.array(np.asarray(x, dtype=np.int64), np.int64(pixels))
