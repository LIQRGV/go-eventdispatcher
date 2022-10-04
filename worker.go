package eventdispatcher

// worker represents the worker that executes the job
type worker struct {
	workerPool chan JobQueue
	jobChannel JobQueue
}

func newWorker(workerPool chan JobQueue) *worker {
	return &worker{
		workerPool: workerPool,
		jobChannel: make(JobQueue, 1),
	}
}

// start method starts the run loop for the worker,
func (w *worker) start() {
	go func() {
		w.workerPool <- w.jobChannel

		for job := range w.jobChannel {
			// re-register the current worker into the worker queue.
			// Freeing up the w.jobChannel
			w.workerPool <- w.jobChannel

			job.Handle()
		}
	}()
}

// stop signals the worker to stop listening for work requests.
func (w *worker) stop() {
	close(w.jobChannel) // workerPool closed on dispatcher already
}
