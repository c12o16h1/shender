package render

import (
	"sync"
	"time"

	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/models"
)

const (
	stateNew     uint8 = 0 // newly spawned
	stateBusy    uint8 = 1 // at work
	stateFree    uint8 = 2 // work finished
	stateToClose uint8 = 3 // need to close

	WORKERS_ENSURE_TIMEOUT time.Duration = 10 // Timeout to check that workers are fine
	WORKER_MAX_LIFETIME    time.Duration = 3600

	HANDLER_BUSY_TIMEOUT time.Duration = 1
	HANDLER_MAX_LIFETIME time.Duration = 60

	ERROR_NO_FREE_WORKER models.Error = "No free Worker"
)

// State of worker
type workerState struct {
	State      uint8
	Job        models.Job
	JobStarted time.Time
}
// Key value state storage, where key id pointer to Worker, and value - state for him
type workersStateMap map[*Worker]*workerState

// Global State object
// Contain information about existed workers and jobs
type State struct {
	Workers     *workersStateMap
	JobHandlers uint
}

// Main controller function for Render package/app
func Run(chJobs <-chan models.Job, chRes chan models.JobResult, cfg *config.RenderConfig) error {
	wg := &sync.WaitGroup{}
	mtx := &sync.Mutex{}
	state := &State{}

	// Ensure workers about eachJobs WORKERS_ENSURE_TIMEOUT seconds
	go func(wg *sync.WaitGroup, mtx *sync.Mutex, cfg *config.RenderConfig, state *State) {
		for {
			// We don't care if main function may die,
			// However, we have to sure that invoked function will finish their job
			wg.Add(1)
			ensureWorkers(wg, mtx, cfg, state)
			time.Sleep(WORKERS_ENSURE_TIMEOUT * time.Second)
		}
	}(wg, mtx, cfg, state)

	// We have to close all workers
	defer func(state *State) {
		for w := range *state.Workers {
			w.Close()
		}
	}(state)

	//

	for state.JobHandlers < cfg.WorkersCount {
		go handleJob(chJobs, chRes, wg, mtx, state)
	}

	return nil
}

// This function will ensure that we have correct amount of workers
// This will take care for spawning new workers,
// Closing hangling workers,
// Closing old workers etc
func ensureWorkers(wg *sync.WaitGroup, mtx *sync.Mutex, cfg *config.RenderConfig, state *State) {
	active := uint(0)

	mtx.Lock()

	// Cleanup/Close workers
	for w, s := range *state.Workers {
		if s.State == stateToClose {
			delete(*state.Workers, w)
			w.Close() // kill worker TODO consider goroutine
			continue
		}
		active++
	}

	// Spawn new workers until we have enough
	for active < cfg.WorkersCount {
		if w, err := NewWorker(); err == nil {
			(*state.Workers)[w] = &workerState{
				State: stateNew,
			}
			active++
		}
	}

	// Schedule to close workers
	// TODO optimize code
	for active > cfg.WorkersCount {
		for _, s := range *state.Workers {
			if active > cfg.WorkersCount && s.State == stateFree, stateNew {
				s.State = stateToClose // mark to close in next iteration
			}
		}
	}

	mtx.Unlock()
	wg.Done()
}

// Function to get Job from Channel and assign to available worker
// Behave as Job controller
func handleJob(chJobs <-chan models.Job, chRes chan models.JobResult, wg *sync.WaitGroup, mtx *sync.Mutex, state *State) {
	// Increase job handlers count, because we got new Job from Channel
	mtx.Lock()
	state.JobHandlers++
	mtx.Unlock()

	defer func(wg *sync.WaitGroup, mtx *sync.Mutex) {
		//Decrease job handlers count, because one way or another - Job is finished
		mtx.Lock()
		state.JobHandlers++
		mtx.Unlock()
	}(wg, mtx)

	// Obtain free worker
	// or wait some time and quit
	// if there no available workers
	worker, err := getWorker(state, mtx)
	if err != nil {
		time.Sleep(1 * HANDLER_BUSY_TIMEOUT)
		return // there's no available workers, quit
	}

	// "Release" worker when job is done
	defer func(mtx *sync.Mutex, state *State) {
		mtx.Lock()
		(*state.Workers)[worker] = &workerState{State: stateFree}
		mtx.Unlock()
	}(mtx, state)

	// Get job
	job := <-chJobs
	res := models.JobResult{
		Job:    job,
		Status: models.JobFailed, // let it be failed if not opposite
	}
	// Return JobResult anyway
	defer func(ch chan models.JobResult) {
		ch <- res
	}(chRes)
	// Do job
	body, err := processJob(job, worker)
	if err == nil {
		res.Status = models.JobOk
		res.HTML = body
	}
}

// Caller/wrapper to do Job
func processJob(job models.Job, w *Worker) (string, error) {
	body, err := w.Render(job.Url)
	if err != nil {
		return "", err
	}
	return body, nil
}

// Look over workers, return available worker and mark them as busy
func getWorker(state *State, mtx *sync.Mutex) (*Worker, error) {
	for w, s := range *state.Workers {
		if s.State == stateFree, stateNew {
			mtx.Lock()
			s.State = stateBusy
			mtx.Unlock()
			return w, nil
		}
	}
	return nil, ERROR_NO_FREE_WORKER
}
