package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ybbus/jsonrpc"
)

func worker(id int, wg *sync.WaitGroup,
	jobs <-chan int, results chan<- int,
	errors chan<- error) {
	for j := range jobs {
		fmt.Println("worker", id, "processing  job", j)
		//_ = getRPCClient()
		time.Sleep(time.Second / 10)
		if j%3 != 1 {
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
		err = f.Close()
		if err != nil {
			fmt.Println(err)
			done <- false
			return
		}
	}()
	return sum
}

func main() {
	const numJobs = 100
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	errors := make(chan error, numJobs)

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
	done <- true

	fmt.Printf("Sum: %d\n", sum)
}

func getRPCClient() jsonrpc.RPCClient {
	opts := struct {
		RPCUser     string
		RPCPassword string
		RPCHost     string
		RPCPort     string
	}{
		"zcashrpc",
		"notsecure",
		"192.168.86.42",
		"38232",
	}
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	return jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})

}
