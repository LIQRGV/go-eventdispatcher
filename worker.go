package eventdispatcher

// worker represents the worker that executes the job
type worker struct {
	workerPool chan JobQueue
	jobChannel JobQueue
}

func newWorker(workerPool chan JobQueue) *worker {
	jobChannel := make(JobQueue)
	workerPool <- jobChannel

	return &worker{
		workerPool: workerPool,
		jobChannel: jobChannel,
	}
}

// start method starts the run loop for the worker,
func (w *worker) start() {
	go func() {
		for job := range w.jobChannel {
			// re-register the current worker into the worker queue.
			// Freeing up the w.jobChannel
			w.workerPool <- w.jobChannel

			job.handle()
		}
	}()
}
