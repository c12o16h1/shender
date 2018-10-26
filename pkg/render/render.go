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

	WORKERS_ENSURE_TIMEOUT time.Duration = 10
	WORKER_MAX_LIFETIME    time.Duration = 3600

	HANDLER_BUSY_TIMEOUT time.Duration = 1
	HANDLER_MAX_LIFETIME time.Duration = 60

	ERROR_NO_FREE_WORKER models.Error = "No free Worker"
)

type State struct {
	Workers     *workersStateMap
	JobHandlers uint
}

func Run(chJobs <-chan models.Job, chRes chan models.JobResult, cfg *config.RenderConfig) error {
	wg := &sync.WaitGroup{}
	mtx := &sync.Mutex{}
	state := &State{}

	// Ensure workers about eachJobs WORKERS_ENSURE_TIMEOUT seconds
	go func(wg *sync.WaitGroup, mtx *sync.Mutex, cfg *config.RenderConfig, state *State) {
		for {
			// Add it there because we don't care if this function may die,
			// However, we have to sure that invoked function will finish their job
			wg.Add(1)
			ensureWorkers(wg, mtx, cfg, state)
			time.Sleep(WORKERS_ENSURE_TIMEOUT * time.Second)
		}
	}(wg, mtx, cfg, state)

	// Cleanup
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

type workersState struct {
	State      uint8
	Job        models.Job
	JobStarted time.Time
}

type workersStateMap map[*Worker]*workersState

// This function will ensure that we have correct amount of workers
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
			(*state.Workers)[w] = &workersState{
				State: stateNew,
			}
			active++
		}
	}

	// Try schedule to close workers until enough
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

func handleJob(chJobs <-chan models.Job, chRes chan models.JobResult, wg *sync.WaitGroup, mtx *sync.Mutex, state *State) {
	// Apply state: Increase job handlers count
	mtx.Lock()
	state.JobHandlers++
	mtx.Unlock()

	defer func(wg *sync.WaitGroup, mtx *sync.Mutex) {
		// Apply state: Decrease job handlers count
		mtx.Lock()
		state.JobHandlers++
		mtx.Unlock()
	}(wg, mtx)

	// Obtain free worker or sleep and quit
	worker, err := getWorker(state, mtx)
	if err != nil {
		time.Sleep(1 * HANDLER_BUSY_TIMEOUT)
	}
	defer func(mtx *sync.Mutex, state *State) {
		mtx.Lock()
		(*state.Workers)[worker] = &workersState{State: stateFree}
		mtx.Unlock()
	}(mtx, state)

	// Get and do job
	job := <-chJobs
	res := models.JobResult{
		Job:    job,
		Status: models.JobFailed, // let it be failed if not opposite
	}
	defer func(ch chan models.JobResult) {
		ch <- res
	}(chRes)
	body, err := processJob(job, worker)
	if err == nil {
		res.Status = models.JobOk
		res.HTML = body
	}
}

func processJob(job models.Job, w *Worker) (string, error) {
	body, err := w.Render(job.Url)
	if err != nil {
		return "", err
	}
	return body, nil
}

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
