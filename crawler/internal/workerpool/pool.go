package workerpool

import (
	"context"
	"sync"
)

// Accumulator is a function type used to aggregate values of type T into a result of type R.
// It must be thread-safe, as multiple goroutines will access the accumulator function concurrently.
// Each worker will produce intermediate results, which are combined with an initial or
// accumulated value.
type Accumulator[T, R any] func(current T, accum R) R

// Transformer is a function type used to transform an element of type T to another type R.
// The function is invoked concurrently by multiple workers, and therefore must be thread-safe
// to ensure data integrity when accessed across multiple goroutines.
// Each worker independently applies the transformer to its own subset of data, and although
// no shared state is expected, the transformer must handle any internal state in a thread-safe
// manner if present.
type Transformer[T, R any] func(current T) R

// Searcher is a function type for exploring data in a hierarchical manner.
// Each call to Searcher takes a parent element of type T and returns a slice of T representing
// its child elements. Since multiple goroutines may call Searcher concurrently, it must be
// thread-safe to ensure consistent results during recursive  exploration.
//
// Important considerations:
//  1. Searcher should be designed to avoid race conditions, particularly if it captures external
//     variables in closures.
//  2. The calling function must handle any state or values in closures, ensuring that
//     captured variables remain consistent throughout recursive or hierarchical search paths.
type Searcher[T any] func(parent T) []T

// Pool is the primary interface for managing worker pools, with support for three main
// operations: Transform, Accumulate, and List. Each operation takes an input channel, applies
// a transformation, accumulation, or list expansion, and returns the respective output.
type Pool[T, R any] interface {
	// Transform applies a transformer function to each item received from the input channel,
	// with results sent to the output channel. Transform operates concurrently, utilizing the
	// specified number of workers. The number of workers must be explicitly defined in the
	// configuration for this function to handle expected workloads effectively.
	// Since multiple workers may call the transformer function concurrently, it must be
	// thread-safe to prevent race conditions or unexpected results when handling shared or
	// internal state. Each worker independently applies the transformer function to its own
	// data subset.
	Transform(ctx context.Context, workers int, input <-chan T, transformer Transformer[T, R]) <-chan R

	// Accumulate applies an accumulator function to the items received from the input channel,
	// with results accumulated and sent to the output channel. The accumulator function must
	// be thread-safe, as multiple workers concurrently update the accumulated result.
	// The output channel will contain intermediate accumulated results as R
	Accumulate(ctx context.Context, workers int, input <-chan T, accumulator Accumulator[T, R]) <-chan R

	// List expands elements based on a searcher function, starting
	// from the given element. The searcher function finds child elements for each parent,
	// allowing exploration in a tree-like structure.
	// The number of workers should be configured based on the workload, ensuring each worker
	// independently processes assigned elements.
	List(ctx context.Context, workers int, start T, searcher Searcher[T])
}

// poolImpl represents Pool implementation
type poolImpl[T, R any] struct{}

// New creates new worker pool
func New[T, R any]() *poolImpl[T, R] {
	return &poolImpl[T, R]{}
}

// Accumulate represents poolImpl implementation of function with the same name
func (p *poolImpl[T, R]) Accumulate(
	ctx context.Context,
	workers int,
	input <-chan T,
	accumulator Accumulator[T, R],
) <-chan R {
	// channel to put accumulated results in
	result := make(chan R)

	// wait group to wait workers to finish their work
	wg := new(sync.WaitGroup)

	for i := 0; i < workers; i++ {
		// implement wait group counter pattern
		wg.Add(1)
		go func() {
			defer wg.Done()
			var res R

			for {
				select {
				// ensure cancelling context is taken into account
				case <-ctx.Done():
					return
				case v, ok := <-input:
					// accumulate result until input channel closes
					if !ok {
						select {
						// ensure cancelling context is taken into account
						case <-ctx.Done():
						case result <- res:
						}
						return
					}

					res = accumulator(v, res)
				}
			}
		}()
	}

	// goroutine for closing result channel when data is in it and results are already accumulated
	go func() {
		defer close(result)
		// wait for all workers to complete
		wg.Wait()
	}()

	return result
}

// List represents poolImpl implementation of function with the same name
func (p *poolImpl[T, R]) List(ctx context.Context, workers int, start T, searcher Searcher[T]) {
	// slice for collecting results on each level
	data := []T{start}

	// iterate over each layer to implement bfs-like tree traverse with synchronisation on
	// each level
	for {
		// if no new data is in data slice then no new layer to process
		if len(data) == 0 {
			return
		}

		// channel from which workers give info to form next level
		input := make(chan T)

		// wait group to wait workers to finish their work
		wg := new(sync.WaitGroup)

		// channel for collecting results on each level
		result := make(chan []T)

		for i := 0; i < workers; i++ {
			// implement wait group counter pattern
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					// ensure cancelling context is taken into account
					case <-ctx.Done():
						return
					case v, ok := <-input:
						if !ok {
							return
						}
						select {
						// ensure cancelling context is taken into account
						case <-ctx.Done():
							return
						case result <- searcher(v):
						}
					}
				}
			}()
		}

		// goroutine for closing result channel when data is in it and results are already searched
		// (it relates only to current level)
		go func() {
			defer close(result)
			// wait for all workers to complete
			wg.Wait()
		}()

		// channel to read data to form new level
		go func() {
			defer close(input)
			for _, v := range data {
				select {
				// ensure cancelling context is taken into account
				case <-ctx.Done():
					return
				case input <- v:
				}
			}
		}()

		// barrier synchronization on current level
		newData := make([]T, 0)
		for {
			select {
			// ensure cancelling context is taken into account
			case <-ctx.Done():
				return
			case r, ok := <-result:
				if !ok {
					// update data when channel is closed and go to next
					// layer
					data = newData
					goto nextIteration
				}
				newData = append(newData, r...)
			}
		}

	nextIteration:
		continue
	}
}

// Transform represents poolImpl implementation of function with the same name
func (p *poolImpl[T, R]) Transform(
	ctx context.Context,
	workers int,
	input <-chan T,
	transformer Transformer[T, R],
) <-chan R {
	// channel for collecting results
	result := make(chan R)

	// wait group to wait workers to finish their work
	wg := new(sync.WaitGroup)

	for i := 0; i < workers; i++ {
		// implement wait group counter pattern
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				// ensure cancelling context is taken into account
				case <-ctx.Done():
					return
				case v, ok := <-input:
					if !ok {
						return
					}

					select {
					// ensure cancelling context is taken into account
					case <-ctx.Done():
						return
					case result <- transformer(v):
					}
				}
			}
		}()
	}

	// goroutine for closing result channel when data is in it and results are
	// already transformed
	go func() {
		defer close(result)
		// wait for all workers to complete
		wg.Wait()
	}()

	return result
}
