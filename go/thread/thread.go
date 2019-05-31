/*package thread contains simple routines for parallelizing generic
tasks. In general, this library should be used as an internal component to
more task-specific parallelization.
*/
package thread

import (
	"fmt"
	"runtime"
)

// Split splits a task up into a specified number of jobs and runs them in
// parallel.
//
// How to use:
//
// jobs := 16
// Split(
//     jobs,
//     func(job int) {
//         /* Do [job] */
//     }
// )
func Split(jobs int, work func(int)) {
	WorkerQueue(jobs, jobs, func(worker, job int) { work(job) })
}

// SplitArrayFunc is the worker funciton for SplitArray. Worker is the worker
// index, and a for loop over the jobs should be written
//     for i := start; i < end; i += step { }
type SplitArrayFunc func(worker, start, end, step int)

type splitArrayConfig struct {
	strategy splitArrayStrategyFlag
	weights []float64
}

type splitArrayStrategyFlag int
const (
	contiguous splitArrayStrategyFlag = iota
	jump
	weightedContiguous
)

// Contiguous causes SplitArray to loop over contiguous chunks of the target
// array. Useful when you want to maintain cache locality.
func Contiguous() splitArrayConfig {
	return splitArrayConfig{ contiguous, nil }
}

// Jump causes SplitArray to range over the whole array with large jumps.
// Useful for load-balancing if there are continguous regions of the array
// which are abnormally expensive to compute.
func Jump() splitArrayConfig {
	return splitArrayConfig{ jump, nil }
}

// Weighted contiguous causes SplitArray to range over contiguous chunks of the
// array that have roughly equal weights. Useful for load-balancing.
func WeightedContiguous(weights []float64) splitArrayConfig {
	return splitArrayConfig{ weightedContiguous, weights }
}

// SplitArray works like Split, but it assumes that Split is being used to
// split work up on an array and handles calculating the for loop indices for
// you. It also takes and optional argument, strategy, that determines how the
// looping works. Call the functions Contiguous(), Jump(), or
// WeightedContiguous(weights) in this argument.
//
// If you want something more complicated, you can build it manually fro
// WorkerQueue().
//
// How to use:
//
// xs := make([]float64, 1000000)
// weights := make([]float64, 1000000)
// workers := 16
//
// SplitArray(
//     workers,
//     func(workter, start, end, step int) {
//         for i := start; i < end; i += step {
//             // Do work on job i.
//         }
//     },
//     WeightedContiguous(weights), // optional
// )
func SplitArray(
	jobs, workers int,
	work SplitArrayFunc,
	config ...splitArrayConfig,
) {
	strat := contiguous
	if len(config) > 0 { strat = config[0].strategy }
	
	switch strat {
	case contiguous:
		splitArrayContiguous(jobs, workers, work)
	case jump:
		splitArrayJump(jobs, workers, work)
	case weightedContiguous:
		splitArrayWeightedContiguous(jobs, workers, config[0].weights, work)
	default:
		panic(fmt.Sprintf("Unknown strategy, %d.", strat))
	}
}

func splitArrayContiguous(jobs, workers int, work SplitArrayFunc) {
	nstep := jobs / workers
	if jobs % workers != 0 { nstep += 1}
	
	Split(
		workers,
		func(worker int) {
			min := worker*nstep
			max := (worker+1)*nstep
			if max > jobs { max = jobs }
			
			work(worker, min, max, 1)
		},
	)
}

func splitArrayJump(jobs, workers int, work SplitArrayFunc) {
	Split(
		workers,
		func(worker int) {
			work(worker, worker, jobs, workers)
		},
	)
}

func splitArrayWeightedContiguous(
	jobs, workers int, weights []float64, work SplitArrayFunc,
) {
	panic("WeightedContiguous not yet implemented.")
}

// WorkerQueue maintains 
//
// How to use:
//
// workers, jobs := 16, 1000
// WorkerQueue(
//     workers, jobs,
//     func(worker, job int) {
//         /* use resources associated with [worker] to do [job] */
//     },
// )
func WorkerQueue(workers, jobs int, work func(worker, job int)) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	jobChan := make(chan int, jobs)
	lockChan := make(chan int, workers)

	for i := 0; i < jobs; i++ { jobChan <- i }
	
	for i := 0; i < workers; i++ {
		go func(workerIdx int) {
			for j := range jobChan {
				work(workerIdx, j)
			}
			lockChan <- 0
		}(i)
	}

	close(jobChan)
	for i := 0; i < workers; i++ { <-lockChan }
}
