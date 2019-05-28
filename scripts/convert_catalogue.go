package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/phil-mansfield/minnow/go/config"
	"github.com/phil-mansfield/minnow/go/minh"
	minnow "github.com/phil-mansfield/minnow/go"
	"github.com/phil-mansfield/minnow/go/text"
	index "github.com/phil-mansfield/minnow/scripts/name_index"
)

type TextConfig struct {
	L, Epsilon float64
	MinParticles int64
	Mp float64
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
	configFname := os.Args[0]
	varsFname := os.Args[1]
	inPattern := os.Args[2]
	out := os.Args[3]

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
	vars.String(&cfg.NameIndex, "NameIndex", "")
	vars.String(&cfg.NameIndex, "TypeIndex", "")

	err := config.ReadConfig(fname, vars)
	if err != nil { panic(err.Error()) }

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
	f := text.OpenRockstar(hlist)
	allNames, blocks, header := f.Names(), f.Blocks(), f.Header()

	buf := []interface{}{ }
	cols := []minh.Column{ }
	names := []string{ }

	for i := range allNames {
		if _, ok := info.Vars[allNames[i]]; !ok { continue }

		t := info.Types[allNames[i]]

		names = append(names, allNames[i])
		if t[0] == "int" {
			buf = append(buf, []int64{ })
			cols = append(cols, minh.Column{Type: minnow.Int64Group})
		} else if t[0] == "float" {
			buf = append(buf, []float32{ })
			cols = append(cols, minh.Column{Type: minnow.Float32Group})
		} else {
			panic(fmt.Sprintf("Type %s not recognized.", t[0]))
		}
	}

	fmt.Println("Buffer:")
	fmt.Println(buf)
	fmt.Println("\nColumns:")
	fmt.Println(cols)
	fmt.Println("\nNames:")
	fmt.Println(names)
	fmt.Println("\nBlocks:")
	fmt.Println(blocks)
	fmt.Println("\nHeader:")
	fmt.Println(header)
}
