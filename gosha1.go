package main

import (
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type result struct {
	Path string
	Sum  []byte
	Size int64
	Err  error
}

type resultSlice []result

func (r resultSlice) Len() int {
	return len(r)
}

func (r resultSlice) Less(i, j int) bool {
	a := &r[i]
	b := &r[j]
	c := bytes.Compare(a.Sum, b.Sum)
	if -1 == c {
		return true
	}
	if 1 == c {
		return false
	}
	c = strings.Compare(a.Path, b.Path)
	if -1 == c {
		return true
	}
	return false
}

func (r resultSlice) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
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

func doSomeJobs(jobs chan string, res chan result, wg *sync.WaitGroup) {
	for path := range jobs {
		sum, size, err := calcSha1(path)
		r := result{path, sum, size, err}
		res <- r
	}
	wg.Done()
}

func produceJobs(dirpath string, jobs chan string, res chan result) {
	err := processDir(dirpath, jobs)
	if err != nil {
		res <- result{"", nil, 0, err}
	}
	close(jobs)
}

func waitForWorkers(wg *sync.WaitGroup, res chan result) {
	wg.Wait()
	close(res)
}

func produceResults(dirpath string) <-chan result {
	n := runtime.NumCPU()
	res := make(chan result)
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

func printResultBuffer(basepath string, rs resultSlice) error {
	var collisions int
	var dupBytes int64
	var dups int
	var size int64
	var sum []byte
	for _, r := range rs {
		p, err := filepath.Rel(basepath, r.Path)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%x\t%s\n", r.Sum, p)
		if !bytes.Equal(r.Sum, sum) {
			sum = r.Sum
			size = r.Size
			continue
		}
		if r.Size != size {
			collisions++
			continue
		}
		dups++
		dupBytes += r.Size
	}
	dupMB := float64(dupBytes) / 1024 / 1024
	fmt.Fprintf(os.Stderr, "Duplicates           : %d\n", dups)
	fmt.Fprintf(os.Stderr, "Duplicate MB         : %f\n", dupMB)
	fmt.Fprintf(os.Stderr, "Collisions (at least): %d\n", collisions)
	return nil
}

func processRootDir(dirpath string) error {
	res := produceResults(dirpath)
	ta := time.Now()
	files := 0
	i := 0
	var MBpsTotal float64
	var bytes int64
	resBuff := make(resultSlice, 0)
	for r := range res {
		bytes += r.Size
		files++
		if r.Err != nil {
			return r.Err
		}
		tb := time.Now()
		s := tb.Sub(ta).Seconds()
		if s > 1.0 {
			i++
			bytesPerSec := float64(bytes) / s
			MBps := bytesPerSec / 1024 / 1024
			MBpsTotal += (MBps - MBpsTotal) / float64(i)
			fmt.Fprintf(os.Stderr, "MB/s: %f\tfiles: %d", MBps, files)
			fmt.Fprintf(os.Stderr, "\tMB/s (total): %f\n", MBpsTotal)
			ta = tb
			bytes = 0
			files = 0
		}
		resBuff = append(resBuff, r)
	}
	sort.Sort(resBuff)
	return printResultBuffer(dirpath, resBuff)
}

func main() {
	flag.Parse()
	dirpath := flag.Arg(0)
	if "" == dirpath {
		fmt.Fprintln(os.Stderr, "ERROR: Arg 0 (dirpath) missing.")
		os.Exit(1)
	}
	err := processRootDir(dirpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: ", err)
		os.Exit(1)
	}
}
