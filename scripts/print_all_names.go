package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/phil-mansfield/minnow/go/text"
)

func main() {
	hlistDir := "/project/surph/rein_dmo/tests/hlists"
	simDirInfo, err := ioutil.ReadDir(hlistDir)
	if err != nil { panic(err.Error()) }

	hlists := make([]string, len(simDirInfo)) 
	for i := range hlists {
		hlists[i] = path.Join(
			hlistDir, simDirInfo[i].Name(), "hlist_1.00000.list",
		)
	}
	
	names := []string{ }

	for i := range hlists {
		if _, err := os.Stat(hlists[i]); err != nil { panic(err.Error())  }
		
		f := text.OpenRockstar(hlists[i])
		names = append(names, f.Names()...)
	}

	for i := range names {
		names[i] = strings.ToLower(names[i])
	}
	sort.Strings(names)

	for i := 0; i < len(names) - 1; i++ {
		if names[i] != names[i + 1] { fmt.Println(names[i+1]) }
	}
}
