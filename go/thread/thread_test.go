package thread

import (
	"testing"
)

func SplitSum(xs []float64, jobs int) float64 {
	// Set up resources for each job to use
	jobSums := make([]float64, jobs)

	segLength  := len(xs) / jobs
	if len(xs) % jobs != 0 { segLength += 1 }
	
	Split(
		jobs,
		// Write callback that uses resouces associated with the given job.
		func(job int) {
			min := job*segLength
			max := (job+1)*segLength
			if max > len(xs) { max = len(xs) }
			for i := min; i < max; i++ {
				jobSums[job] += xs[i]
			}
		},
	)

	// combine results
	sum := 0.0
	for i := range jobSums { sum += jobSums[i] }

	return sum
}

func TestSplit(t *testing.T) {
	xs := make([]float64, 1000)
	for i := range xs { xs[i] = 1 }
	
	tests := []int{1, 2, 3, 49, 100, 1000}
	for i := range tests {
		sum := SplitSum(xs, tests[i])
		
		if sum != float64(len(xs)) {
			t.Errorf("Sum(workers = %d) = %g, not %d", tests[i], sum, len(xs))
		}
	}
}

func SplitArrayContiguousSum(xs []float64, jobs int) float64 {
	jobSums := make([]float64, jobs)
	
	SplitArray(
		len(xs), jobs,
		func(worker, istart, iend, istep int) {
			for i := istart; i < iend; i += istep {
				jobSums[worker] += xs[i]
			}
		},
		Contiguous(),
	)

	sum := 0.0
	for i := range jobSums { sum += jobSums[i] }

	return sum
}

func TestSplitArrayContiguous(t *testing.T) {
	xs := make([]float64, 1000)
	for i := range xs { xs[i] = 1 }
	
	tests := []int{1, 2, 3, 49, 100, 1000}
	for i := range tests {
		sum := SplitArrayContiguousSum(xs, tests[i])
		
		if sum != float64(len(xs)) {
			t.Errorf("Sum(workers = %d) = %g, not %d", tests[i], sum, len(xs))
		}
	}
}

func SplitArrayJumpSum(xs []float64, jobs int) float64 {
	jobSums := make([]float64, jobs)
	
	SplitArray(
		len(xs), jobs,
		func(worker, istart, iend, istep int) {
			for i := istart; i < iend; i += istep {
				jobSums[worker] += xs[i]
			}
		},
		Jump(),
	)

	sum := 0.0
	for i := range jobSums { sum += jobSums[i] }

	return sum
}

func TestSplitArrayJump(t *testing.T) {
	xs := make([]float64, 1000)
	for i := range xs { xs[i] = 1 }
	
	tests := []int{1, 2, 3, 49, 100, 1000}
	for i := range tests {
		sum := SplitArrayJumpSum(xs, tests[i])
		
		if sum != float64(len(xs)) {
			t.Errorf("Sum(workers = %d) = %g, not %d", tests[i], sum, len(xs))
		}
	}
}
