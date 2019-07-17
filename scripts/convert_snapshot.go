package main

import (
	"fmt"

	"github.com/phil-mansfield/minnow/go/minp"
)

func main() {
	inType inDir := os.Args[1], os.Args[2]
	outType, outDir := os.Args[3], os.Args[4]

	var (
		snap snapshot.Snapshot
		err error
	)
	switch inType {
	case "lgadget-2":
		snap, err = snapshot.NewLGadget2(inDir)
		if err != nil { panic(err.Error()) }
	case "minp":
		snap, err = snapshot.NewMinP(inDir)
		if err != nil { panic(err.Error()) }
	default:
		panic(fmt.Sprintf("Unrecognized"))
	}

	lGadget2Header := snapshot.BytesToLGadget2Header(snap.RawHeader(0))

	switch outType {
	case "lgadget-2":
		fnameFmt := os.Args[5]
		err := snapshot.WriteLGadget2(dir, fnameFmt, snap, lGadget2Header)
		if err != nil { panic(err.Error()) }
	case "minp":
		fnameFmt := os.Args[5]
		fileCellsStr, subCellsStr := os.Args[6], os.Args[7]
		dxStr, dvStr := os.Args[8], os.Args[9]

		fileCells, err := strconv.Atoi(fileCellsStr)
		if err != nil { panic(err.Error()) }
		subCells, err := strconv.Atoi(subCellsStr)
		if err != nil { panic(err.Error()) }

		dx, err := strconv.ParseFloat(dxStr, 64)
		if err != nil { panic(err.Error()) }
		dv, err := strconv.ParseFloat(dvStr, 64)
		if err != nil { panic(err.Error()) }

		err := snapshot.WriteMinP(
			dir, fnameFmt, snap, fileCells, subCells, dx, dv,
		)
		if err != nil { panic(err.Error()) }
	}
}
