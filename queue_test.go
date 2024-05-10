package testqueue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type InMemoryProcessor struct {
	jobs []Job
}

func (processor *InMemoryProcessor) Process(jobs []Job) []string {
	time.Sleep(300 * time.Millisecond)
	processor.jobs = append(processor.jobs, jobs...)
	responses := make([]string, len(jobs))
	for index := range responses {
		responses[index] = fmt.Sprintf("response: %d", index)
	}

	return responses
}

func TestDoNotProcessUntilFullQueue(t *testing.T) {
	processor := &InMemoryProcessor{jobs: make([]Job, 0)}
	queue := CreateQueue(3, 2*time.Second, processor)

	job1 := Job{Id: 1}
	result1 := queue.Process(job1)
	assert.Equal(t, result1.Job(), job1)
	assert.True(t, isChannelOpen(result1.Ready()))
	assert.Equal(t, "", result1.Content())

	queue.Process(Job{Id: 2})
	assert.Equal(t, 0, len(processor.jobs))
	assert.True(t, isChannelOpen(result1.Ready()))

	job3 := Job{Id: 3}
	result3 := queue.Process(job3)

	select {
	case <-result3.Ready():
		break
	case <-time.After(310 * time.Millisecond):
		t.Log("expected to handle exactly 300 ms")
		t.Fail()
	}
	assert.Equal(t, 3, len(processor.jobs))
	assert.Equal(t, result3.Job(), job3)

	assert.Equal(t, "response: 0", result1.Content())
	assert.Equal(t, "response: 2", result3.Content())
}

func TestProcessNotFullQueueBySchedule(t *testing.T) {
	processor := &InMemoryProcessor{jobs: make([]Job, 0)}
	queue := CreateQueue(3, 250*time.Millisecond, processor)

	job1 := Job{Id: 1}
	result1 := queue.Process(job1)
	assert.Equal(t, result1.Job(), job1)
	assert.True(t, isChannelOpen(result1.Ready()))
	assert.Equal(t, "", result1.Content())

	job2 := Job{Id: 2}
	result2 := queue.Process(job2)
	assert.Equal(t, 0, len(processor.jobs))
	assert.Equal(t, result2.Job(), job2)
	assert.True(t, isChannelOpen(result1.Ready()))
	assert.True(t, isChannelOpen(result2.Ready()))
	assert.Equal(t, "", result1.Content())
	assert.Equal(t, "", result1.Content())

	waitCycles := 0
out:
	for {
		select {
		case <-time.After(270 * time.Millisecond):
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
	assert.Equal(t, 2, len(processor.jobs))

	select {
	case <-result2.Ready():
		break
	case <-time.After(260 * time.Millisecond):
		t.Log("expected to handle exactly 1 second")
		t.Fail()
	}
	assert.Equal(t, 2, len(processor.jobs))

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
