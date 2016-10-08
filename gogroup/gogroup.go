/**
 * GoGroup - A tiny helper for using goroutines with a WaitGroup.
 */
package gogroup

import "sync"

/**
 * Starts a group of goroutines and calls done when all of them are done.
 * 
 * @param numWorkers  Number of goroutines to start.
 * @param work        The function to start as goroutine(s).
 * @param done        The function to call when all workers are done.
 */
func Go(numWorkers int, work, done func()) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			work()
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		done()
	}()
}

