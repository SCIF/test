package testqueue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type InMemoryProcessor struct {
	jobs           []Job
	processingTime time.Duration
}

func (processor *InMemoryProcessor) Process(jobs []Job) []string {
	processor.jobs = append(processor.jobs, jobs...)
	responses := make([]string, len(jobs))
	for index := range responses {
		responses[index] = fmt.Sprintf("response: %d", index)
	}

	time.Sleep(processor.processingTime)

	return responses
}

func TestDoNotProcessUntilFullQueue(t *testing.T) {
	processor := &InMemoryProcessor{jobs: make([]Job, 0), processingTime: 100 * time.Millisecond}
	queue := NewQueue(3, 2*time.Second, processor)

	job1 := Job{Id: 1}
	result1 := queue.Process(job1)

	assert.Equal(t, result1.Job(), job1)
	assertJobResultIsNotHandledYet(t, result1)

	result2 := queue.Process(Job{Id: 2})

	assertJobResultIsNotHandledYet(t, result1)
	assertJobResultIsNotHandledYet(t, result2)

	result3 := queue.Process(Job{Id: 3})

	// all jobs must be sent immediately to processor. allow 1ms just for make sure goroutine is executed
	time.Sleep(time.Millisecond)
	assert.Equal(t, 3, len(processor.jobs))

	select {
	// check readyness channel works and reacts
	case <-result3.Ready():
		break
	case <-time.After(110 * time.Millisecond):
		t.Log("expected to be handle within the processor processingTime")
		t.Fail()
	}

	// once ready, check the content is in place
	assert.Equal(t, "response: 0", result1.Content())
	assert.Equal(t, "response: 1", result2.Content())
	assert.Equal(t, "response: 2", result3.Content())
}

func TestProcessNotFullQueueBySchedule(t *testing.T) {
	processor := &InMemoryProcessor{jobs: make([]Job, 0), processingTime: 300 * time.Millisecond}
	queue := NewQueue(3, 250*time.Millisecond, processor)

	job1 := Job{Id: 1}
	result1 := queue.Process(job1)
	assert.Equal(t, result1.Job(), job1)
	assertJobResultIsNotHandledYet(t, result1)

	result2 := queue.Process(Job{Id: 2})

	// make sure nothing is sent to the processor yet
	assert.Equal(t, 0, len(processor.jobs))
	assertJobResultIsNotHandledYet(t, result1)
	assertJobResultIsNotHandledYet(t, result2)

	// make sure queue delay is met
	waitCycles := 0
	failTimeout := time.After(260 * time.Millisecond) // high boundary check
out:
	for {
		select {
		case <-failTimeout:
			t.Log("expected to handle faster than this timeout")
			t.Fail()
			break out
		case <-time.Tick(51 * time.Millisecond):
			waitCycles++
			if 0 < len(processor.jobs) {
				t.Log("Break happen")
				break out
			}
		}
	}

	assert.Equal(t, 5, waitCycles)
	// check messages sent
	assert.Equal(t, 2, len(processor.jobs))

	// check message result is resolved once batchProcessor returned result
	select {
	case <-result2.Ready():
		break
	case <-time.After(310 * time.Millisecond):
		t.Log("expected to handle faster than timeout")
		t.Fail()
	}

	assert.Equal(t, "response: 0", result1.Content())
	assert.Equal(t, "response: 1", result2.Content())
}

func isChannelOpen(channel chan struct{}) bool {
	select {
	case <-channel:
		return false
	default:
		return true
	}
}

func assertJobResultIsNotHandledYet(t *testing.T, result JobResult) {
	assert.True(t, isChannelOpen(result.Ready()))
	assert.Equal(t, "", result.Content())
}
