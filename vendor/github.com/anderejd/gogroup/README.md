# gogroup
GoGroup - A tiny helper for using goroutines with a WaitGroup.

Example usage, fan-out workers, fan-in results:
```
/** 
 * The returned Result channel will close when done to allow the Result consumer
 * to iterate on the Result channel using a range loop.
 */
func produceResults() <-chan Result {
	rs := make(chan Result)
	js := make(chan Job)
	work := func() {
		for j := range js {
			rs <- transformJobToResult(j)
		}
	}
	gogroup.Go(runtime.NumCPU(), work, func() { close(rs) })
	go produceJobs(js)
	return rs
}

func consumeResults(rs <-chan Result) {
	for r := range rs {
		...
	}
}
```
