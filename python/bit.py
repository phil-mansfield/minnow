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
