package workerqueue

const (
	totalWorkers = 10
)

type Work func(id int)

type Inbound chan Work

type WorkerQueue chan Inbound

type Worker struct {
	id          int
	inbound     Inbound
	workerQueue WorkerQueue
	quit        chan bool
	done        chan bool
}

func NewWorker(id int, workerQueue WorkerQueue) Worker {
	worker := Worker{
		id:          id,
		inbound:     make(chan Work),
		workerQueue: workerQueue,
		quit:        make(chan bool),
		done:        make(chan bool),
	}
	return worker
}

func (w *Worker) Start() {
	go func() {
		for {
			// we are ready for work!
			w.workerQueue <- w.inbound
			select {
			case work := <-w.inbound:
				work(w.id)
			case <-w.quit:
				w.done <- true
				return
			}
		}
	}()
}

// Tells the worker to stop once it has finished it's current job.
func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

// Waits for a worker to finish before returning.
func (w *Worker) WaitForFinish() {
	<-w.done
}

// func NewJob(i int) Work {
// 	return func(id int) {
// 		time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
// 		fmt.Printf("%d: working... task# %d\n", id, i)
// 	}
// }

// ////
// func main() {
// 	// create the workerqueue
// 	workerQueue := make(WorkerQueue, totalWorkers)

// 	// create and start the workers
// 	workers := make([]Worker, totalWorkers)
// 	for i := range workers {
// 		workers[i] = NewWorker(i, workerQueue)
// 		workers[i].Start()
// 	}

// 	// add some jobs
// 	for i := 0; i < 100; i++ {
// 		// get a free worker
// 		worker := <-workerQueue
// 		// give it some work
// 		worker <- NewJob(i)
// 	}

// 	// tell the workers to stop waiting for work
// 	for i := range workers {
// 		workers[i].Stop()
// 	}

// 	// wait for the workers to quit
// 	for i := range workers {
// 		workers[i].WaitForFinish()
// 	}
// }
