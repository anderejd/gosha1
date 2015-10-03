package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func ProcessFile(path string) error {
	return nil
}

func ProcessDir(path string) error {
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
			err = ProcessFile(p)
			if nil != err {
				return err
			}
			continue
		}
		err = ProcessDir(p)
		if nil != err {
			return err
		}
		fmt.Println(p)
	}
	return nil
}

func work(in chan string, out chan int) {
	for s := range c {
		job <- c
		out <- job
	}
}

func createWorkers(n int, in chan string, out chan int ) {
	for i := 0; i < n; i++ {
		go work(in, out)
	}
}

func main() {
	flag.Parse()
	dirpath := flag.Arg(0) if "" == dirpath {
		fmt.Fprintln(os.Stderr, "Arg 0 (dirpath) missing.")
		os.Exit(1)
	}
	jobs := make(chan string)
	results := make(chan int)
	numWorkers := runtime.NumCPU()
	createWorkers(numWorkers, jobs, results)
	err := ProcessDir(dirpath, jobs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
	for res := range
}
