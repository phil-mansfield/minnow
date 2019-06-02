package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"runtime"

	"github.com/phil-mansfield/minnow/go/minh"
)

func main() {
	cells, err := strconv.Atoi(os.Args[1])
	if err != nil { panic(err.Error()) }

	bnd, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil { panic(err.Error()) }

	inPattern := os.Args[3]
	out := os.Args[4]

	minhFiles, err := filepath.Glob(inPattern)
	if err != nil { panic(err.Error()) }

	for _, fname := range minhFiles {
		fmt.Println("Conerting", fname)

		t0 := time.Now()
		ConvertFile(fname, OutName(out, fname), cells, bnd)
		t1 := time.Now()
		dt := t1.Sub(t0)

		fmt.Printf("    %.2f minutes\n", dt.Seconds() / 60)
	}
}

func OutName(outDir, hlist string) string {
	base := path.Base(hlist)
	tok := strings.Split(base, ".")
	if len(tok) == 0 {
		tok = append(tok, "bnd.minh")
	} else {
		tok = append(tok[:len(tok) - 1], "bnd.minh")
	}
	return path.Join(outDir, strings.Join(tok, "."))
}

func ConvertFile(inName, outName string, cells int, bnd float64) {
	in := minh.Open(inName)
	out := minh.CreateBoundary(outName)
	defer in.Close()
	defer out.Close()

	out.Header(in.Text)
	out.Geometry(in.L, float32(bnd), cells)
	
	coord := in.Floats([]string{"x", "y", "z"})
	out.Coordinates(coord["x"], coord["y"], coord["z"])

	for i := range in.Names {
		runtime.GC()

		var data interface{}

		switch in.Columns[i].Type {
		case minh.Float, minh.Float32:
			data = in.Floats([]string{in.Names[i]})[in.Names[i]]
		case minh.Int, minh.Int64:
			data = in.Ints([]string{in.Names[i]})[in.Names[i]]
		}
		
		out.Column(in.Names[i], in.Columns[i], data)
	}
}
