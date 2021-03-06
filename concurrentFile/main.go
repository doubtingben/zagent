package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup,
	jobs <-chan int, results chan<- int,
	errors chan<- error) {
	for j := range jobs {
		fmt.Println("worker", id, "processing  job", j)
		time.Sleep(time.Second / 10)
		if j != 7 {
			results <- j
		} else {
			errors <- fmt.Errorf("error on job %v", j)
		}
		wg.Done()
	}
}

func collectResults(results chan int, errors chan error, done chan bool) (sum int) {
	go func() {
		f, err := os.Create("concurrent")
		if err != nil {
			fmt.Println(err)
			return
		}
		for {
			select {
			case <-done:
				return
			case err := <-errors:
				fmt.Println("ERROR:", err.Error())
			case result := <-results:
				fmt.Println("FINISHED:", result)
				_, err = fmt.Fprintln(f, result)
				if err != nil {
					fmt.Println(err)
					f.Close()
					done <- false
					return
				}
			}
		}
	}()
	return sum
}

func main() {
	const numJobs = 20
	jobs := make(chan int, numJobs)
	results := make(chan int)
	errors := make(chan error)

	var wg sync.WaitGroup
	for w := 1; w <= 10; w++ {
		go worker(w, &wg, jobs, results, errors)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
		wg.Add(1)
	}
	close(jobs)

	done := make(chan bool)
	sum := collectResults(results, errors, done)

	wg.Wait()
	close(results)
	done <- true

	fmt.Printf("Sum: %d", sum)
}
