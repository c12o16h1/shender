package broker

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os/exec"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/c12o16h1/shender/pkg/models"
)

const (
	MAX_CPU_LOAD     float64 = 60 // Max acceptable load for CPU
	MAX_MEMORY_USAGE float64 = 75 // Max acceptable usage of memory

	MAIN_LOOP_TIMEOUT = 10 * time.Millisecond // Timeout in main process loop to let CPU do more important things

	MAX_RENDERERS = 10 // Max amount of alive workers

	// Port range to run renderer workers
	MIN_RENDEDER_PORT int = 52500
	MAX_RENDEDER_PORT int = 57750

	ERR_INVALID_WORKER = models.Error("Invalid worker")
)

var (
	busyPorts = make(map[int]bool)
	port      = MIN_RENDEDER_PORT
)

/*
Crawler crawl websites pages and get cache from them
 */
func Crawl(chJobs <-chan models.Job, chRes chan<- models.JobResult) error {
	var wg sync.WaitGroup
	var mtx sync.Mutex
	limiter := make(chan struct{}, MAX_RENDERERS)
	go func() {
		for {
			log.Print("LIMITER:", len(limiter), cap(limiter))
			log.Print("JOBS:", len(chJobs), cap(chJobs))
			log.Print("JOBRES:", len(chRes), cap(chRes))
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		if !enoughResources() {
			time.Sleep(MAIN_LOOP_TIMEOUT) // Sleep a bit, let CPU do other, more important loops
		}
		limiter <- struct{}{}         // would block if we already have enough renderers
		time.Sleep(MAIN_LOOP_TIMEOUT) // Sleep a bit, let CPU do other, more important loops
		job := <-chJobs

		mtx.Lock()
		port = nextPort(port)
		mtx.Unlock()

		go spawnRenderer(port) // lifetime of rendeder is max 30 seconds, so it's safe
		go handleJob(mtx, wg, job, chRes, port, limiter)
	}
	return nil
}

// Goroutine to take job and port for worker, and process them
func handleJob(mtx sync.Mutex, wg sync.WaitGroup, j models.Job, chRes chan<- models.JobResult, port int, limiter <-chan struct{}) {
	wg.Add(1)
	result := models.JobResult{
		Status: models.JobFailed,
		Job:    j,
	}
	defer func() {
		log.Print(port, ":", result.Job.Url, " : ", len(result.HTML))
		chRes <- result
		wg.Done()
		<-limiter // Drain from limiter channel, so allowing to spawn new workers and do other jobs
		// Free port
		mtx.Lock()
		delete(busyPorts, port)
		mtx.Unlock()
	}()

	time.Sleep(1 * time.Second) // wait for renderer

	// Create worker for this task
	c, err := rpc.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Print(0, err)
		return
	}
	defer c.Close()
	// Check that worker is ok
	var ok string
	err = c.Call("Worker.Heartbeat", "", &ok)
	if err != nil || ok != models.OK {
		log.Print(1, err)
		return
	}
	// Do job
	url := "http://" + j.Url
	log.Print("ENQ:", port, ":", url)
	err = c.Call("Worker.Render", url, &result.HTML)
	if err != nil {
		log.Print(2, port, err)
		return
	}
	// Close worker, kill chrome etc
	c.Call("Worker.Close", 0, nil)

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
func nextPort(current int) int {
	if current < MIN_RENDEDER_PORT {
		current = MIN_RENDEDER_PORT
	}
	p := current

	for {
		p += 2
		if current < MAX_RENDEDER_PORT {
			if _, ok := busyPorts[p]; !ok {
				// Check that port is free
				if !isFreePort(p) {
					continue
				}
				// Check that chrome ports is free
				for _, cp := range chromePorts(p) {
					if !isFreePort(cp) {
						continue
					}
				}
				// Check that chrome ports is free
				busyPorts[p] = true
				return p
			}
		}

	}
	log.Print("Can't allocate PORT")
	return MIN_RENDEDER_PORT
}

// Spawn render worker at specified port
func spawnRenderer(port int) error {
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

// Chrome may take 2 ports.
// It'll be from -10000 to -9999
func chromePorts(port int) []int {
	return []int{
		port - 10000,
		port - 9999,
	}
}


// Checks that port is free
// True for free port
func isFreePort(p int) bool {
	conn, _ := net.Dial("tcp", net.JoinHostPort("127.0.0.1", string(p)))
	if conn != nil {
		conn.Close()
		return false
	}
	return true
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
