package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Result struct {
	Path string
	Sum  []byte
	Size int64
	Err  error
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
	h := sha1.New()
	written, err = io.Copy(h, f)
	if nil != err {
		return
	}
	sum = h.Sum(nil)
	return
}

func doSomeJobs(jobs chan string, res chan Result, wg *sync.WaitGroup) {
	for path := range jobs {
		sum, size, err := calcSha1(path)
		r := Result{path, sum, size, err}
		res <- r
	}
	wg.Done()
}

func produceJobs(dirpath string, jobs chan string, res chan Result) {
	err := processDir(dirpath, jobs)
	if err != nil {
		res <- Result{"", nil, 0, err}
	}
	close(jobs)
}

func waitForWorkers(wg *sync.WaitGroup, res chan Result) {
	wg.Wait()
	close(res)
}

func produceResults(dirpath string) <-chan Result {
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

func main() {
	flag.Parse()
	dirpath := flag.Arg(0)
	if "" == dirpath {
		fmt.Fprintln(os.Stderr, "Arg 0 (dirpath) missing.")
		os.Exit(1)
	}
	res := produceResults(dirpath)
	ta := time.Now()
	files := 0
	i := 0
	var MBpsTotal float64 = 0
	var bytes int64 = 0
	for r := range res {
		bytes += r.Size
		files++
		if r.Err != nil {
			fmt.Fprintln(os.Stderr, r.Err)
			break
		}
		tb := time.Now()
		s := tb.Sub(ta).Seconds()
		if s > 1.0 {
			i++
			bytesPerSec := float64(bytes) / s
			MBps := bytesPerSec / 1024 / 1024
			MBpsTotal += (MBps - MBpsTotal) / float64(i)
			fmt.Printf("MB/s: %f\tfiles: %d", MBps, files)
			fmt.Printf("\tMB/s (total): %f\n", MBpsTotal)
			ta = tb
			bytes = 0
			files = 0
		}
		//fmt.Printf("%x\t%d\t%s\n", r.Sum, r.Size, r.Path)
	}
}
