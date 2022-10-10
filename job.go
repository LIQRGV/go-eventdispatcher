package eventdispatcher

import (
	"log"
	"runtime/debug"
)

// JobQueue A channel that we can send work requests on.
type JobQueue chan Job

func NewJobQueue(maxQueue int) JobQueue {
	if maxQueue > 0 {
		return make(JobQueue, maxQueue)
	}

	return make(JobQueue)
}

type Job interface {
	Handle()
}

// DefaultJob represents the job to be run
type DefaultJob struct {
	function  func()
	panicFunc func(byte []byte)
}

func defaultPanicFunc(byte []byte) {
	defaultLog := log.Default()
	defaultLog.Print(string(byte))
}

func NewDefaultJob(f func(), f2 ...func(byte []byte)) *DefaultJob {
	panicFunc := defaultPanicFunc

	if len(f2) > 1 {
		panic("panic function must 1 or not exists")
	}

	if len(f2) == 1 {
		panicFunc = f2[0]
	}

	return &DefaultJob{
		function:  f,
		panicFunc: panicFunc,
	}
}

func (j *DefaultJob) Handle() {
	defer func() {
		if r := recover(); r != nil {
			j.panicFunc(debug.Stack())
		}
	}()

	j.function()
}
