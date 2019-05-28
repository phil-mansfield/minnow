package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phil-mansfield/minnow/go/config"
	"github.com/phil-mansfield/minnow/go/minh"
	"github.com/phil-mansfield/minnow/go/text"
	index "github.com/phil-mansfield/minnow/scripts/name_index"
)

type TextConfig struct {
	L float64
	Epsilon float64
	MinParticles int64
	Mp float64
	MassName string
	NameIndex string
	TypeIndex string
}

type FileInfo struct {
	Config *TextConfig
	Index *index.Index
	Vars map[string]bool
	Types map[string][]string
}

func main() {
	configFname := os.Args[1]
	varsFname := os.Args[2]
	inPattern := os.Args[3]
	out := os.Args[4]

	cfg := ParseConfig(configFname)
	idx := index.Open(cfg.NameIndex)
	types := ParseTypes(cfg.TypeIndex, idx)
	vars := ParseVars(varsFname, idx)

	hlistFiles, err := filepath.Glob(inPattern)
	if err != nil { panic(err.Error()) }

	for i := range hlistFiles {
		hlist := hlistFiles[i]
		hlistOut := path.Join(out, path.Base(hlist))
		info := &FileInfo{ Config: cfg, Index: idx, Vars: vars, Types: types }
		ConvertFile(info, hlist, hlistOut)
	}
}

func ParseConfig(fname string) *TextConfig {
	cfg := &TextConfig{ }

	vars := config.NewConfigVars("minh")
	vars.Float(&cfg.L, "L", 0)
	vars.Float(&cfg.Epsilon, "Epsilon", 0)
	vars.Float(&cfg.Mp, "Mp", 0)
	vars.Int(&cfg.MinParticles, "MinParticles", 0)
	vars.String(&cfg.MassName, "MassName", "mvir")
	vars.String(&cfg.NameIndex, "NameIndex", "")
	vars.String(&cfg.TypeIndex, "TypeIndex", "")

	err := config.ReadConfig(fname, vars)
	if err != nil { panic(err.Error()) }

	if cfg.L == 0 { panic(fmt.Sprintf("L not set in %s", fname)) }
	if cfg.Epsilon == 0 { panic(fmt.Sprintf("Epsilon not set in %s", fname)) }
	if cfg.Mp == 0 { panic(fmt.Sprintf("Mp not set in %s", fname)) }
	if cfg.MinParticles == 0 {
		panic(fmt.Sprintf("MinParticles not set in %s", fname))
	}
	if cfg.NameIndex == "" {
		panic(fmt.Sprintf("NameIndex not set in %s", fname))
	}
	if cfg.TypeIndex == "" {
		panic(fmt.Sprintf("TypeIndex not set in %s", fname))
	}

	return cfg
}

func ParseTypes(fname string, idx *index.Index) map[string][]string {
	bText, err := ioutil.ReadFile(fname)
	if err != nil { panic(err.Error()) }
	tok := clean(strings.Split(string(bText), "\n"))

	out := map[string][]string{ }
	for _, line := range tok {
		words := clean(strings.Split(line, " "))
		v, typeInfo := words[0], words[1:]
		std, ok := idx.Standardize(v)
		if !ok {
			panic(fmt.Sprintf("Variable %s isn't contained in name index.", v))
		}

		out[std] = typeInfo
	}

	return out
}

func ParseVars(fname string, idx *index.Index) map[string]bool {
	bText, err := ioutil.ReadFile(fname)
	if err != nil { panic(err.Error()) }
	tok := clean(strings.Split(string(bText), " "))

	out := map[string]bool{ }
	for _, v := range tok {
		std, ok := idx.Standardize(v)
		if !ok {
			panic(fmt.Sprintf("Variable %s isn't contained in name index.", v))
		}
		out[std] = true
	}

	return out
}

func clean(tok []string) []string {
	for i := range tok {
		tok[i] = strings.Trim(tok[i], " \n\t")
	}

	out := []string{}
	for i := range tok {
		if len(tok[i]) == 0 { continue }
		out = append(out, tok[i])
	}

	return out
}

func ConvertFile(info *FileInfo, hlist, out string) {
	fR := text.OpenRockstar(hlist)
	allNames, header := fR.Names(), fR.Header()

	buf := []interface{}{ }
	cols := []minh.Column{ }
	names := []string{ }

	// Type parsing.
	for i := range allNames {
		std, ok := info.Index.Standardize(allNames[i])
		allNames[i] = std
		if !ok {
			panic(fmt.Sprintf("Column name %s from %s not in name index",
				allNames[i], hlist))
		}
		if _, ok := info.Vars[std]; !ok { continue }

		t := info.Types[std]

		names = append(names, std)
		buf, cols = ParseTypeString(info.Config, buf, cols, t)
	}

	// Cutoff nonsense
	cutoff := float32(info.Config.Mp * float64(info.Config.MinParticles))
	iMass := find(names, info.Config.MassName)
	if iMass == -1 {
		panic(fmt.Sprintf("MassName %s not in name index.",
			info.Config.MassName))
	}

	// Actual I/O

	fR.SetNames(allNames)

	fM := minh.Create(out)
	fM.Header(names, header, cols)
	for b := 0; b < fR.Blocks(); b++ {
		fmt.Println(b, names)
		fR.Block(b, names, buf)
		n := GenericCut(cutoff, buf[iMass], buf)
		fmt.Println(n)
		if n > 0 { fM.Block(buf) }
	}
	fmt.Println("Closing")
	fM.Close()
}

func find(names []string, name string) int {
	for i := range names {
		if name == names[i] { return i }
	}
	return -1
}

func ParseTypeString(
	cfg *TextConfig, buf []interface{},
	cols []minh.Column, t []string,
) ([]interface{}, []minh.Column) {
	switch t[0] {
	case "int":
		fmt.Println("int")
		buf = append(buf, []int64{ })
		cols = append(cols, minh.Column{Type: minh.Int})
	case "q_float":
		fmt.Println("float")
		col := minh.Column{Type: minh.Float}
		switch t[1] {
		case "position":
			col.Low, col.High, col.Dx = 0, float32(cfg.L), float32(cfg.Epsilon)
		case "log":
			min, err := strconv.ParseFloat(t[2], 64)
			if err != nil { panic(err.Error())  }
			max, err := strconv.ParseFloat(t[3], 64)
			if err != nil { panic(err.Error())  }
			eps, err := strconv.ParseFloat(t[4], 64)
			if err != nil { panic(err.Error())  }

			col.Log, col.Dx = 1, float32(eps)
			col.Low = float32(math.Log10(min))
			col.High = float32(math.Log10(max))
		default:
			panic(fmt.Sprintf("q_float qualifier %s not recognized", t[1]))
		}
		
		buf = append(buf, []float32{ })
		cols = append(cols, col)
	default:
		panic(fmt.Sprintf("Type %s not recognized.", t[0]))
	}
	fmt.Println(cols[len(cols) - 1])
	return buf, cols
}

func GenericCut(cutoff float32, mass interface{}, buf []interface{}) int {
	m := mass.([]float32)
	ok, n := make([]bool, len(m)), 0
	for i := range ok {
		ok[i] = m[i] > cutoff
		if ok[i] { n++ }
	}

	for i := range buf {
		switch x := buf[i].(type) {
		case []float32: buf[i] = filterFloat32(x, ok)
		case []int64: buf[i] = filterInt64(x, ok)
		default: panic(fmt.Sprintf("Unknown type %T in GenericCut", x))
		}
	}

	return n
}

func filterFloat32(x []float32, ok []bool) []float32 {
	j := 0
	for i := range x {
		if ok[i] {
			x[j] = x[i]
			j++
		}
	}
	x = x[:j]
	return x
}

func filterInt64(x []int64, ok []bool) []int64 {
	j := 0
	for i := range x {
		if ok[i] {
			x[j] = x[i]
			j++
		}
	}
	x = x[:j]
	return x
}
