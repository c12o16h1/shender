package broker

import (
	"bytes"
	"fmt"
	"net/rpc"
	"os/exec"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/pkg/errors"
)

const (
	MAX_CPU_LOAD     float64 = 60 // Max acceptable load for CPU
	MAX_MEMORY_USAGE float64 = 75 // Max acceptable usage of memory

	MAIN_LOOP_TIMEOUT = 10 * time.Millisecond // Timeout in main process loop to let CPU do more important things

	MAX_RENDERERS = 10 // Max amount of alive workers

	// Port range to run renderer workers
	MIN_RENDEDER_PORT uint = 53500
	MAX_RENDEDER_PORT uint = 53600

	ERR_INVALID_WORKER = models.Error("Invalid worker")
)

func Run(chJobs <-chan models.Job, chRes chan models.JobResult) error {
	var wg sync.WaitGroup
	port := MIN_RENDEDER_PORT
	limiter := make(chan struct{}, MAX_RENDERERS)

	for enoughResources() {
		limiter <- struct{}{}         // would block if we already have enough renderers
		time.Sleep(MAIN_LOOP_TIMEOUT) // Sleep a bit, let CPU do other, more important loops
		job := <-chJobs
		port = nextPort(port)
		if err := spawnRenderer(port); err != nil {
			continue
		}
		go handleJob(wg, job, chRes, port, limiter)
	}
	return nil
}

// Goroutine to take job and port for worker, and process them
func handleJob(wg sync.WaitGroup, job models.Job, chRes chan models.JobResult, port uint, limiter <-chan struct{}) {
	wg.Add(1)
	result := models.JobResult{
		Status: models.JobFailed,
	}
	defer func() {
		chRes <- result
		wg.Done()
		<-limiter // Drain from limiter channel, so allowing to spawn new workers and do other jobs
	}()

	// Create worker for this task
	c, err := rpc.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		errors.Wrap(err, ERR_INVALID_WORKER.Error())
		return
	}
	// Check that worker is ok
	var ok string
	err = c.Call("Worker.Hearbeat", nil, &ok)
	if err != nil || ok != models.OK {
		fmt.Println(err)
		return
	}
	// Do job
	err = c.Call("Worker.Render", job.Url, &result.HTML)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Close worker, kill chrome etc
	err = c.Call("Worker.Close", nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	result.Status = models.JobOk
}

// Func to check that we have enough CPU and memory to do something,
// f.e. spawn new workers or start new jobs
func enoughResources() bool {
	m, err := mem.VirtualMemory()
	if err != nil {
		return false
	}
	c, err := cpu.Percent(0, false)
	if err != nil {
		return false
	}

	if m.UsedPercent < MAX_MEMORY_USAGE && c[0] < MAX_CPU_LOAD {
		return true
	}

	return false
}

// Round robin algorithm for ports
func nextPort(current uint) uint {
	if current <= MAX_RENDEDER_PORT {
		current++
		return current
	}
	return MIN_RENDEDER_PORT
}

// Spawn render worker at specified port
func spawnRenderer(port uint) error {
	args := []string{"-port", fmt.Sprintf("%d", port)}
	cmd := exec.Command("./bin/render", args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return nil
}

func sampleIcomingQueue(chJobs chan models.Job) {
	jobs := []models.Job{
		{
			Url: "http://google.com",
		},
		{
			Url: "http://react.com",
		},
		{
			Url: "http://angular.io",
		},
	}

	for {
		for _, j := range jobs {
			time.Sleep(5 * time.Second)
			chJobs <- j
		}
	}

}
