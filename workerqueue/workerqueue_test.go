package workerqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkerWithoutWorkerQueue(t *testing.T) {
	var wq WorkerQueue
	_, err := NewWorker(1, wq)
	assert.Error(t, err)
}

func TestNewWorker(t *testing.T) {
	wq := make(WorkerQueue, 1)
	worker, err := NewWorker(1, wq)
	assert.NoError(t, err)
	assert.Equal(t, 1, worker.id)
	assert.Equal(t, wq, worker.workerQueue)
}

func TestWorkerStartStopAndWaitForFinish(t *testing.T) {
	wq := make(WorkerQueue, 1)
	worker, _ := NewWorker(1, wq)
	worker.Start()

	expected := <-wq
	assert.Equal(t, expected, worker.inbound)

	for {
		select {
		case <-time.After(10 * time.Second):
			assert.Fail(t, "Cannot stop worker")
			return
		default:
			worker.Stop()
			worker.WaitForFinish()
			return
		}
	}
}

func TestWorkerJob(t *testing.T) {
	wq := make(WorkerQueue, 1)
	worker, _ := NewWorker(1, wq)
	worker.Start()

	w := <-wq

	out := make(chan int)

	// run job
	w <- func(id int) {
		assert.Equal(t, 1, id)
		out <- 12345
	}

	// check worker
FINISH:
	for {
		select {
		case result := <-out:
			assert.Equal(t, 12345, result)
			break FINISH
		case <-time.After(10 * time.Second):
			assert.Fail(t, "Fail to receive result from the workerqueue")
			return
		}
	}

	// stop worker
	for {
		select {
		case <-time.After(10 * time.Second):
			assert.Fail(t, "Cannot stop worker")
			return
		default:
			worker.Stop()
			worker.WaitForFinish()
			return
		}
	}
}
