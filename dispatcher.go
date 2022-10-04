package eventdispatcher

import "sync"

type Dispatcher struct {
	maxWorkers int

	// A pool of workers channels that are registered with the dispatcher
	workerPool chan JobQueue
	queueChan  JobQueue

	quit      chan bool
	waitGroup *sync.WaitGroup
}

func NewDispatcher(maxWorkers int, queueChan JobQueue) *Dispatcher {
	var wg sync.WaitGroup
	return &Dispatcher{
		maxWorkers: maxWorkers,

		workerPool: make(chan JobQueue, maxWorkers),
		queueChan:  queueChan,

		quit:      make(chan bool),
		waitGroup: &wg,
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		workerObj := newWorker(d.workerPool)

		workerObj.start()
		d.waitGroup.Add(1)

		go func(localWorker *worker) {
			defer d.waitGroup.Done()

			<-d.quit
			localWorker.stop()
		}(workerObj)
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
		d.quit <- true
	}

	d.waitGroup.Wait()

	// stopping n number of workers
	<-d.workerPool
	close(d.workerPool)
}
