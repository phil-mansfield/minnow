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
import time

palette.configure(False)

def main():
    fname = sys.argv[1]
    f = minh.open(fname)

    for b in range(f.blocks):
        plt.figure()

        t0 = time.time()
        bnd, x, y, z, mvir = f.block(
            b, ["boundary", "x", "y", "z", "mvir"]
        )
        t1 = time.time()
        print("Read block %d: %.2f minutes" % (b, (t1 - t0)/60))

        ok = (z < 25) & (mvir > 1e12)

        plt.plot(x[ok & (bnd == 0)], y[ok & (bnd == 0)], ".", c="r")
        plt.plot(x[ok & (bnd == 1)], y[ok & (bnd == 1)], ".", c="k")

        plt.xlim(0, 125)
        plt.ylim(0, 125)
        plt.xlabel(r"$X$")
        plt.ylabel(r"$Y$")

        plt.savefig("slice_b%d.png" % b)

if __name__ == "__main__": main()
