package eventdispatcher_test

import (
	"context"
	"fmt"
	"github.com/LIQRGV/eventdispatcher"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestEventDispatcher_should_process_any_job(t *testing.T) {
	maxQueue := 1
	maxWorkers := 3
	jobQueue := eventdispatcher.NewJobQueue(maxQueue)
	dispatcher := eventdispatcher.NewDispatcher(
		maxWorkers,
		jobQueue,
	)
	dispatcher.Run()
	defer dispatcher.Stop()

	randWithSource := rand.NewSource(time.Now().UnixNano())
	random := rand.New(randWithSource)
	jobNumber := random.Intn(10) + 1
	testChan := make(chan int)

	for i := 0; i < jobNumber; i++ {
		jobFunc := func(currentNum int) func() {
			return func() {
				testChan <- currentNum
			}
		}(i)

		jobQueue <- eventdispatcher.NewDefaultJob(jobFunc)
	}

	testCounter := 0

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	go func(cancelFunc context.CancelFunc) {
		for range testChan {
			testCounter += 1

			if testCounter == jobNumber {
				break
			}
		}

		cancelFunc()
	}(cancel)

	<-ctx.Done()

	switch ctx.Err() {
	case context.DeadlineExceeded:
		assert.Fail(t, "Context Deadline Exceeded before execution done")

	case context.Canceled:
		assert.Equal(t, jobNumber, testCounter, fmt.Sprintf("jobNumber %d is not equal testCounter %d", jobNumber, testCounter))
	}
}

func TestEventDispatcher_should_not_quit_upon_panic(t *testing.T) {
	maxQueue := 1
	maxWorkers := 1
	jobQueue := eventdispatcher.NewJobQueue(maxQueue)
	dispatcher := eventdispatcher.NewDispatcher(
		maxWorkers,
		jobQueue,
	)

	dispatcher.Run()
	defer dispatcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	jobFunc := func() {
		cancel() // this is odd but works
		panic("me panic!!!")
	}
	jobQueue <- eventdispatcher.NewDefaultJob(jobFunc)

	<-ctx.Done()

	switch ctx.Err() {
	case context.DeadlineExceeded:
		assert.Fail(t, "Context Deadline Exceeded before execution done")

	case context.Canceled:
		assert.True(t, true, "process should finished despite panic happened")
	}
}

func TestJobQueue_able_to_make_unbuffered_queue(t *testing.T) {
	jobQueue := eventdispatcher.NewJobQueue(0) // 0 means unbuffered

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	go func() {
		jobQueue <- eventdispatcher.NewDefaultJob(func() {})

		assert.Fail(t, "This line should unreachable")
		cancel()
	}()

	<-ctx.Done()

	isNeverConsumed := false

	switch ctx.Err() {
	case context.DeadlineExceeded:
		isNeverConsumed = true
	}

	assert.True(t, isNeverConsumed, "The queue should never be consumed")
}

func TestJobQueue_able_to_receive_custom_panicFunc(t *testing.T) {
	maxQueue := 1
	maxWorkers := 3
	jobQueue := eventdispatcher.NewJobQueue(maxQueue)
	dispatcher := eventdispatcher.NewDispatcher(
		maxWorkers,
		jobQueue,
	)
	dispatcher.Run()
	defer dispatcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	panicReceivedFlag := false

	panicFunc := func(b []byte) {
		panicReceivedFlag = true
		cancel()
	}

	jobQueue <- eventdispatcher.NewDefaultJob(func() {
		panic("me panic!!!")
	}, panicFunc)

	<-ctx.Done()

	switch ctx.Err() {
	case context.DeadlineExceeded:
		assert.Fail(t, "Context Deadline Exceeded before execution done")
	}

	assert.True(t, panicReceivedFlag, "The panic should triggered")
}

func TestJobQueue_unable_to_receive_more_than_1_custom_panicFunc(t *testing.T) {
	maxQueue := 1
	jobQueue := eventdispatcher.NewJobQueue(maxQueue)

	panicFunc := func(b []byte) {}

	defer func() {
		recoverFlag := false
		if r := recover(); r != nil {
			recoverFlag = true
		}

		assert.True(t, recoverFlag, "Panic-Recover should triggered")
	}()

	jobQueue <- eventdispatcher.NewDefaultJob(func() {
		panic("me panic!!!")
	}, panicFunc, panicFunc)
}
