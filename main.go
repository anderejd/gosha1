package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func ProcessFile(path string) (error) {
	return nil
}

func ProcessDir(path string) (error) {
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

func main() {
	flag.Parse()
	dirpath := flag.Arg(0)
	if "" == dirpath {
		fmt.Fprintln(os.Stderr, "Arg 0 (dirpath) missing.")
		os.Exit(1)
	}
	err := ProcessDir(dirpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}
