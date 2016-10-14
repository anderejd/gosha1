/**
 * A tiny helper for using goroutines with a WaitGroup.
 */
package syncext

import "sync"

/**
 * Starts and waits for a group of goroutines. This function is blocking if the
 * callback is nil.
 *
 * @param numWorkers  Number of goroutines to start.
 * @param work        The function to start as goroutine(s).
 * @param done        The function to call when all workers are done. If this is
 *                    nil this function will block until all workers have
 *                    finished.
 */
func FanOut(numWorkers int, work, done func()) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			work()
			wg.Done()
		}()
	}
	if nil == done {
		wg.Wait()
		return
	}
	go func() {
		wg.Wait()
		done()
	}()
}
