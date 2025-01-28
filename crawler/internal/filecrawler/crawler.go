package crawler

import (
	"context"
	"crawler/internal/fs"
	"crawler/internal/workerpool"
	"encoding/json"
	"sync"
)

// Configuration holds the configuration for the crawler, specifying the number of workers for
// file searching, processing, and accumulating tasks. The values for SearchWorkers, FileWorkers,
// and AccumulatorWorkers are critical to efficient performance and must be defined in
// every configuration.
type Configuration struct {
	SearchWorkers      int // Number of workers responsible for searching files.
	FileWorkers        int // Number of workers for processing individual files.
	AccumulatorWorkers int // Number of workers for accumulating results.
}

// Combiner is a function type that defines how to combine two values of type R into a single
// result. Combiner is not required to be thread-safe
//
// Combiner can either:
//   - Modify one of its input arguments to include the result of the other and return it,
//     or
//   - Create a new combined result based on the inputs and return it.
//
// It is assumed that type R has a neutral element (forming a monoid)
type Combiner[R any] func(current R, accum R) R

// Crawler represents a concurrent crawler implementing a map-reduce model with multiple workers
// to manage file processing, transformation, and accumulation tasks. The crawler is designed to
// handle large sets of files efficiently, assuming that all files can fit into memory
// simultaneously.
type Crawler[T, R any] interface {
	// Collect performs the full crawling operation, coordinating with the file system
	// and worker pool to process files and accumulate results. The result type R is assumed
	// to be a monoid, meaning there exists a neutral element for combination, and that
	// R supports an associative combiner operation.
	// The result of this collection process, after all reductions, is returned as type R.
	//
	// Important requirements:
	// 1. Number of workers in the Configuration is mandatory for managing workload efficiently.
	// 2. FileSystem and Accumulator must be thread-safe.
	// 3. Combiner does not need to be thread-safe.
	// 4. If an accumulator or combiner function modifies one of its arguments,
	//    it should return that modified value rather than creating a new one,
	//    or alternatively, it can create and return a new combined result.
	// 5. Context cancellation is respected across workers.
	// 6. Type T is derived by json-deserializing the file contents, and any issues in deserialization
	//    must be handled within the worker.
	// 7. The combiner function will wait for all workers to complete, ensuring no goroutine leaks
	//    occur during the process.
	Collect(
		ctx context.Context,
		fileSystem fs.FileSystem,
		root string,
		conf Configuration,
		accumulator workerpool.Accumulator[T, R],
		combiner Combiner[R],
	) (R, error)
}

// crawlerImpl represents Crawler implementation
type crawlerImpl[T, R any] struct{}

// New creates new crawler
func New[T, R any]() *crawlerImpl[T, R] {
	return &crawlerImpl[T, R]{}
}

// fileStorage serves for handling concurrent access to files
type fileStorage struct {
	fileMu map[string]*sync.Mutex
	mu     *sync.RWMutex
}

// newFileStorage initializes fileStorage with sync.RWMutex to allow concurrent reading of its contents
// and maps each filename to sync.Mutex to allow sequential reading of file by multiple goroutines
func newFileStorage() *fileStorage {
	return &fileStorage{
		fileMu: make(map[string]*sync.Mutex),
		mu:     new(sync.RWMutex),
	}
}

// protect wraps given function to recover from panics while saving an error
func protect[T any](aE *atomicErr, fn func(string) T) func(string) T {
	return func(arg string) (result T) {
		defer func() {
			if err := recover(); err != nil {
				// here it is expected that err is a standard error
				if e, ok := err.(error); ok {
					aE.addError(e)
				}
			}
		}()

		// call a function before we will recover from panic
		return fn(arg)
	}
}

// atomicErr serves to protect error from concurrent access from multiple goroutines
type atomicErr struct {
	err error
	mu  *sync.Mutex
}

// addError saves error to atomicErr if it hasn't been written
func (a *atomicErr) addError(e error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.err == nil {
		a.err = e
	}
}

// Collect represents crawlerImpl implementation of function with the same name
func (c *crawlerImpl[T, R]) Collect(
	ctx context.Context,
	fileSystem fs.FileSystem,
	root string,
	conf Configuration,
	accumulator workerpool.Accumulator[T, R],
	combiner Combiner[R],
) (R, error) {
	// channel required to start pipeline by sending names of searched files to it
	fileChan := make(chan string)

	// Each worker pool serves to work with a certain stage of file system processing
	searchWp := workerpool.New[string, string]()
	transformWp := workerpool.New[string, T]()
	resultWp := workerpool.New[T, R]()

	fStorage := newFileStorage()

	// wait group to ensure no additional work is needed to write to file channel
	listWg := sync.WaitGroup{}

	aE := &atomicErr{
		mu: new(sync.Mutex),
	}

	listWg.Add(1)
	go func() {
		defer listWg.Done()
		searchWp.List(ctx, conf.SearchWorkers, root, protect(aE, func(parent string) []string {
			listWg.Add(1)
			defer listWg.Done()

			// get dir entries
			dirEntries, err := fileSystem.ReadDir(parent)
			if err != nil {
				aE.addError(err)
				return nil
			}

			// directories traversal
			var dirs []string
			for _, entry := range dirEntries {
				join := fileSystem.Join(parent, entry.Name())
				// check dir entry type
				if entry.IsDir() {
					dirs = append(dirs, join)
				} else {
					select {
					// ensure cancelling context is taken into account
					case <-ctx.Done():
						return nil
					case fileChan <- join:
					}
				}
			}
			return dirs
		}))
	}()

	// wait group to ensure goroutine closing file channel is finished
	fWg := sync.WaitGroup{}

	// fWg.Done guarantees that file channel is closed and working goroutines are finished
	fWg.Add(1)
	go func() {
		// closing channel to stop pipeline
		defer close(fileChan)
		listWg.Wait()
		fWg.Done()
	}()

	// at this stage files are read, deserialized and their results are sent to type channel
	typeCh := transformWp.Transform(ctx, conf.FileWorkers, fileChan, protect(aE, func(current string) T {
		f, err := fileSystem.Open(current)

		defer func() {
			_ = f.Close()
		}()

		var result T

		if err != nil {
			aE.addError(err)
			return result
		}

		// such a buffer size is enough to make one read
		const bufferSize = 512
		var content []byte
		buffer := make([]byte, bufferSize)

		fStorage.mu.RLock()
		// allow readers to read file content
		fMu, exists := fStorage.fileMu[current]
		fStorage.mu.RUnlock()

		// if there is no data yet then one reader should become a writer
		if !exists {
			fStorage.mu.Lock()
			fMu, exists = fStorage.fileMu[current]
			// the mutex could have already been created during the waiting time
			if !exists {
				fMu = new(sync.Mutex)
				fStorage.fileMu[current] = fMu
			}
			fStorage.mu.Unlock()
		}
		// everyone who wants to read a file will read it
		fMu.Lock()
		defer fMu.Unlock()

		// one read to buffer is enough in this implementation
		n, readErr := f.Read(buffer)
		content = buffer[:n]

		if readErr != nil {
			aE.addError(readErr)
			return result
		}

		// deserialize file content
		er := json.Unmarshal(content, &result)
		if er != nil {
			aE.addError(er)
			return result
		}

		return result
	}))

	// apply accumulator function to deserialized values from files
	resultCh := resultWp.Accumulate(ctx, conf.AccumulatorWorkers, typeCh, accumulator)

	var result R

	// this slice serves to collect values from result channel allowing combiner to wait
	// for pipeline completion
	var resultValues []R

	for {
		res, ok := <-resultCh
		if !ok {
			// at the moment when the channel is closed there will be no
			// simultaneous writing and reading of aE.err
			if aE.err != nil {
				return result, aE.err
			}

			// wait for file channel to close
			fWg.Wait()
			// at this stage the combiner waited for the pipeline to finish working
			for _, rv := range resultValues {
				result = combiner(rv, result)
			}
			return result, ctx.Err()
		}

		// while the channel with the results is open they are not processed
		resultValues = append(resultValues, res)
	}
}
