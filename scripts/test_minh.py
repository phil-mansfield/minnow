from __future__ import division, print_function
import sys
sys.path.append("../python")

import matplotlib
matplotlib.use("PDF")

import numpy as np
import matplotlib.pyplot as plt
import palette
from palette import pc
import minh

palette.configure(False)

def main():
    fname = sys.argv[1]
    f = minh.open(fname)

    for b in range(f.blocks):
        x, y, z, mvir, pid = f.block(b, ["x", "y", "z", "mvir", "pid"])

        ok = (mvir > 1e13) & (pid == -1) & (z < 25)
        c = pc()
        plt.plot(x[ok], y[ok], ".", c=c)

    plt.xlim(0, 125)
    plt.ylim(0, 125)

    plt.savefig("slice.pdf")

if __name__ == "__main__": main()
