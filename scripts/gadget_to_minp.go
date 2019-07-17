package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/phil-mansfield/minnow/go/minp"
	"github.com/phil-mansfield/minnow/go/minp/snapshot"
)

func main() {
	inDir, outDir := os.Args[1], os.Args[2]
	fileCellsStr, subCellsStr := os.Args[3], os.Args[4]
	dxStr, dvStr := os.Args[5], os.Args[6]
	
	fileCells, err := strconv.Atoi(fileCellsStr)
	if err != nil { panic(err.Error()) }
	subCells, err := strconv.Atoi(subCellsStr)
	if err != nil { panic(err.Error()) }
	dx, err := strconv.ParseFloat(dxStr, 64)
	if err != nil { panic(err.Error()) }
	dv, err := strconv.ParseFloat(dvStr, 64)
	if err != nil { panic(err.Error()) }

	snap, err := snapshot.LGadget2(inDir)
	if err != nil { panic(err.Error()) }
	snap = snapshot.NewGrid(snap, fileCells)

	for i := 0; i < snap.Files(); i++ {
		c := minp.Cell{ int64(i), int64(fileCells), int64(subCells) }
		f := minp.Create(fmt.Sprintf("%s/x.%03d.minp", outDir, i))

		f.Header(snap.Header(), snap.RawHeader(i), c, dx, true)

		x, err := snap.ReadX(i)
		if err != nil { panic(err.Error())  }
		f.Vectors(x)

		f.Close()
	}

	for i := 0; i < snap.Files(); i++ {
		c := minp.Cell{ int64(i), int64(fileCells), int64(subCells) }
		f := minp.Create(fmt.Sprintf("%s/v.%03d.minp", outDir, i))

		f.Header(snap.Header(), snap.RawHeader(i), c, dv, false) 

		v, err := snap.ReadV(i)
		if err != nil { panic(err.Error())  }
		f.Vectors(v)

		f.Close()
	}
}
