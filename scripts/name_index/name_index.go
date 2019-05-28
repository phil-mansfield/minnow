package name_index

import (
	"io/ioutil"
	"strings"
	"sort"
)

type Index struct {
	allNames []string
	orig, standard []string
}

func (idx *Index) Len() int { return len(idx.orig) }
func (idx *Index) Less(i, j int) bool { return idx.orig[i] < idx.orig[j] }
func (idx *Index) Swap(i, j int) {
	idx.orig[i], idx.orig[j] = idx.orig[j], idx.orig[i]
	idx.standard[i], idx.standard[j] = idx.standard[j], idx.standard[i]
}

func Open(fname string) *Index {
	bText, err := ioutil.ReadFile(fname)
	if err != nil { panic(err.Error()) }
	text := string(bText)
	
	lines := clean(strings.Split(text, "\n"))

	orig, standard, all := []string{}, []string{}, []string{}
	for i := range lines {
		tok := clean(strings.Split(lines[i], " "))
		
		all = append(all, tok[0])
		for j := range tok {
			orig = append(orig, tok[j])
			standard = append(standard, tok[0])
		}
	}

	idx := &Index{ orig: orig, standard: standard, allNames: all }
	sort.Sort(idx)

	return idx
}

func (idx *Index) Standardize(name string) (std string, inIndex bool) {
	name = strings.ToLower(name)
	i := sort.SearchStrings(idx.orig, name)

	if i == len(idx.orig) || idx.orig[i] != name {
		return name, false
	} else {
		return idx.standard[i], true
	}
}

func (idx *Index) AllNames() []string { return idx.allNames }

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
