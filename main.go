package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/anderejd/gosha1/sha1dir"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type resultSlice []sha1dir.Result

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

func printResultBuffer(basepath string, rs resultSlice) error {
	var collisions int
	var dupBytes int64
	var totBytes int64
	var dups int
	var size int64
	var sum []byte
	for _, r := range rs {
		p, err := filepath.Rel(basepath, r.Path)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%x\t%s\n", r.Sum, p)
		totBytes += r.Size
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
	totMB := float64(totBytes) / 1024 / 1024
	log("Duplicates           : ", dups)
	log("Duplicate MB         : ", dupMB)
	log("Total MB             : ", totMB)
	//log("Collisions (at least): ", collisions)
	return nil
}

func log(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func logStatus(MBps float64, files int, MBpsTotal float64) {
	const format = "MB/s: %.2f\tfiles: %d\tMB/s (total): %.2f\n"
	fmt.Fprintf(os.Stderr, format, MBps, files, MBpsTotal)
}

func processRootDir(dirpath string) error {
	res := sha1dir.ProduceConcurrent(dirpath)
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
			logStatus(MBps, files, MBpsTotal)
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
		log("ERROR: Arg 0 (dirpath) missing.")
		os.Exit(1)
	}
	err := processRootDir(dirpath)
	if err != nil {
		log("ERROR: ", err)
		os.Exit(1)
	}
}
