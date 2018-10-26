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

	JOB_MAX_LIFETIME time.Duration = 60
)

type State struct {
	Workers *workersStateMap
}

func Run(ch <-chan []models.Job, cfg *config.RenderConfig) error {
	wg := &sync.WaitGroup{}
	mtx := &sync.Mutex{}
	state := &State{}

	// Ensure workers about each WORKERS_ENSURE_TIMEOUT seconds
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

	for
	cfg.WorkersCount

	for i := uint(0); i < cfg.WorkersCount; i++ {
		go Render("https://www.contextu.com")
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

func handleJobs(wg *sync.WaitGroup, mtx *sync.Mutex, cfg *config.RenderConfig, state *State) {

}
