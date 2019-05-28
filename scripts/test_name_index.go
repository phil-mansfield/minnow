package main

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/phil-mansfield/minnow/go/text"
	index "github.com/phil-mansfield/minnow/scripts/name_index"
)

func main() {
	nameIndex := "name_index/name_index.txt"
	hlistDir := "/project/surph/rein_dmo/tests/hlists"

	idx := index.Open(nameIndex)
	
	simDirInfo, err := ioutil.ReadDir(hlistDir)
	if err != nil { panic(err.Error()) }
	hlists := make([]string, len(simDirInfo)) 
	for i := range hlists {
		hlists[i] = path.Join(
			hlistDir, simDirInfo[i].Name(), "hlist_1.00000.list",
		)
	}

	for i := range hlists {
		names := text.OpenRockstar(hlists[i]).Names()

		for j := range names  {
			if _, ok := idx.Standardize(names[j]); !ok {
				fmt.Println(names[j])
			}
		}
	}
}
