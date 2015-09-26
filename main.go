package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	flag.Parse()
	path := flag.Arg(0)
	if "" == path {
		fmt.Fprintln(os.Stderr, "Arg 0 (path) missing.")
		os.Exit(1)
	}
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fmt.Println(f.Name())
	}
}
