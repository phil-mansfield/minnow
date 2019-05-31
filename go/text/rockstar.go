package text

import (
	"strings"
)

type Rockstar struct {
	rd *Reader
}

func OpenRockstar(fname string, configOpt ...ReaderConfig) *Rockstar {
	r := &Rockstar{ rd: Open(fname, configOpt...) }
	return r
}

func (r *Rockstar) Names() []string {
	firstLine := r.rd.LineHeader(1)
	if strings.Contains(firstLine, "(0)") {
		return consistentTreesNames(firstLine)
	} else {
		return rockstarNames(firstLine)
	}
}

func rockstarNames(line string) []string {
	names := strings.Split(line[1:], " ")
	for i := range names {
		names[i] = strings.Trim(names[i], " \n\t")
	}
	return names
}

func consistentTreesNames(line string) []string {
	names := rockstarNames(line)
	for i := range names {
		toks := strings.Split(names[i], "(")
		names[i] = strings.Join(toks[:len(toks)-1], "(")
	}

	return names
}

func (r *Rockstar) SetThreads(n int) { r.rd.SetThreads(n) }
func (r *Rockstar) Header() string { return r.rd.CommentHeader() }
func (r *Rockstar) SetNames(names []string) { r.rd.SetNames(names) }
func (r *Rockstar) Blocks() int { return r.rd.Blocks() }
func (r *Rockstar) Close() { r.rd.Close() }
func (r *Rockstar) Block(b int, names []string, out []interface{}) {
	r.rd.Block(b, names, out)
}
