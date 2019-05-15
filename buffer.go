package minnow

func sizeInt64Buf(buf []int64, n int) []int64 {
	if n <= cap(buf) { return buf[:n] }
	buf = buf[:cap(buf)]
	buf = append(buf, make([]int64, n - cap(buf))...)

	return buf
}
