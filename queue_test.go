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
	queue := CreateQueue(3, 250, processor)

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
		t.Log("expected to handle exactly 1 second")
		t.Fail()
	}
	assert.Equal(t, 3, len(processor.jobs))
	assert.Equal(t, result3.Job(), job3)

	assert.Equal(t, "response: 0", result1.Content())
	assert.Equal(t, "response: 2", result3.Content())
}

func isChannelOpen(channel chan struct{}) bool {
	select {
	case <-channel:
		return false
	default:
		return true
	}
}
