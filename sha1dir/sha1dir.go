package sha1dir

import (
	"crypto/sha1"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Result struct for a single file.
// Err will be nil on success.
type Result struct {
	Path string
	Sum  []byte
	Size int64
	Err  error
}

// ProduceConcurrent calculates the SHA1 sum for each file in the directory
// tree and returns the results over the returned channel. The channel will be
// closed when all files have been processed.
func ProduceConcurrent(dirpath string) <-chan Result {
	n := runtime.NumCPU()
	res := make(chan Result)
	jobs := make(chan string)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go doSomeJobs(jobs, res, &wg)
	}
	go waitForWorkers(&wg, res)
	go produceJobs(dirpath, jobs, res)
	return res
}

func isDotPath(p string) bool {
	b := filepath.Base(p)
	if ".." != b && len(b) > 1 && '.' == b[0] {
		return true
	}
	return false
}

func processDir(path string, jobs chan<- string) error {
	if isDotPath(path) {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return err
	}
	for _, f := range list {
		p := filepath.Join(path, f.Name())
		if !f.IsDir() {
			if f.Mode().IsRegular() {
				jobs <- p
			}
		} else {
			err = processDir(p, jobs)
			if nil != err {
				return err
			}
		}
	}
	return nil
}

func calcSha1(path string) (sum []byte, written int64, err error) {
	var f *os.File
	f, err = os.Open(path)
	if nil != err {
		return
	}
	defer f.Close()
	h := sha1.New()
	written, err = io.Copy(h, f)
	if nil != err {
		return
	}
	sum = h.Sum(nil)
	return
}

func doSomeJobs(jobs <-chan string, res chan<- Result, wg *sync.WaitGroup) {
	for path := range jobs {
		sum, size, err := calcSha1(path)
		r := Result{path, sum, size, err}
		res <- r
	}
	wg.Done()
}

func produceJobs(dirpath string, jobs chan<- string, res chan<- Result) {
	err := processDir(dirpath, jobs)
	if err != nil {
		res <- Result{"", nil, 0, err}
	}
	close(jobs)
}

func waitForWorkers(wg *sync.WaitGroup, res chan<- Result) {
	wg.Wait()
	close(res)
}
