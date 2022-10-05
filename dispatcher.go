package eventdispatcher

type Dispatcher struct {
	maxWorkers int

	// A pool of workers channels that are registered with the dispatcher
	workerPool chan JobQueue
	queueChan  JobQueue

	quit chan bool
}

func NewDispatcher(maxWorkers int, queueChan JobQueue) *Dispatcher {
	return &Dispatcher{
		maxWorkers: maxWorkers,

		workerPool: make(chan JobQueue, maxWorkers),
		queueChan:  queueChan,

		quit: make(chan bool),
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		workerObj := newWorker(d.workerPool)

		workerObj.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for job := range d.queueChan {
		// a job request has been received
		go func(job Job) {
			// try to obtain a worker job channel that is available.
			// this will block until a worker is idle
			jobChannel := <-d.workerPool

			// dispatch the job to the worker job channel
			jobChannel <- job
		}(job)
	}
}

func (d *Dispatcher) Stop() {
	for i := 0; i < d.maxWorkers; i++ {
		close(<-d.workerPool)
	}

	close(d.workerPool)
}
