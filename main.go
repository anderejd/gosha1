package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type Result struct {
	err  error
	path string
}

func processDir(path string, jobs chan string) error {
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
			jobs <- p
			continue
		}
		err = processDir(p, jobs)
		if nil != err {
			return err
		}
	}
	return nil
}

func work(jobs chan string, res chan Result, wg *sync.WaitGroup) {
	for j := range jobs {
		r := Result{nil, j}
		res <- r
	}
	wg.Done()
}

func produceJobs(dirpath string, jobs chan string, res chan Result) {
	err := processDir(dirpath, jobs)
	if err != nil {
		res <- Result{err, ""}
	}
	close(jobs)
}

func waitForWorkers(wg *sync.WaitGroup, res chan Result) {
	wg.Wait()
	close(res)
}

func produceResults(dirpath string) <-chan Result {
	res := make(chan Result)
	jobs := make(chan string)
	n := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go work(jobs, res, &wg)
	}
	go waitForWorkers(&wg, res)
	go produceJobs(dirpath, jobs, res)
	return res
}

func main() {
	flag.Parse()
	dirpath := flag.Arg(0)
	if "" == dirpath {
		fmt.Fprintln(os.Stderr, "Arg 0 (dirpath) missing.")
		os.Exit(1)
	}
	res := produceResults(dirpath)
	for r := range res {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, r.err)
			break
		}
	}
}
