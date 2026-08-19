package workers

import (
	"errors"
	"log"

	"github.com/grqphical/f-stop/internal/database"
	"github.com/grqphical/f-stop/internal/models"
	"github.com/grqphical/f-stop/internal/storage"
)

const maxRetries int = 5

type WorkerFunction = func(models.JobPayload, database.DBInterface, storage.StorageInterface) error

type WorkerManager struct {
	workerCount      int
	workerFunction   WorkerFunction
	shutdownChannels []chan bool
	db               database.DBInterface
	si               storage.StorageInterface
}

func NewWorkerManager(workerCount int, workerFunction WorkerFunction, db database.DBInterface, si storage.StorageInterface) *WorkerManager {
	shutdownChannels := make([]chan bool, workerCount)

	wm := &WorkerManager{
		workerCount,
		workerFunction,
		shutdownChannels,
		db,
		si,
	}

	for i := range workerCount {
		shutdownChan := make(chan bool, 1)
		wm.shutdownChannels[i] = shutdownChan

		go wm.WorkerRunner(i)
		log.Printf("[WORKER %d] Ready.\n", i)
	}

	return wm
}

func (wm *WorkerManager) Close() {
	for i := range wm.workerCount {
		wm.shutdownChannels[i] <- true
	}
}

func (wm *WorkerManager) WorkerRunner(id int) {
	shutdownChan := wm.shutdownChannels[id]

	for {
		select {
		case <-shutdownChan:
			return
		default:

		}

		job, err := wm.db.DequeueJob(1, maxRetries)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				continue
			}
			log.Printf("[WORKER %d] (ERROR): %v\n", id, err)
			continue
		}
		log.Printf("[WORKER %d] Dequeued job with ID %d\n", id, job.ID)

		err = wm.workerFunction(job.Payload, wm.db, wm.si)
		if err != nil {
			log.Printf("[WORKER %d] (ERROR): %v\n", id, err)
			wm.db.AcknowledgeFailure(job.ID)
		} else {
			wm.db.AcknowledgeSuccess(job.ID)
		}
	}
}
