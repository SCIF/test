package testqueue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type InMemoryProcessor struct {
	jobs []Job
}

func (processor *InMemoryProcessor) Process(jobs []Job) []string {
	processor.jobs = append(processor.jobs, jobs...)
	responses := make([]string, len(jobs))
	for index, _ := range responses {
		responses[index] = fmt.Sprintf("response: %d", index)
	}

	return responses
}

func TestDoNotProcessUntilFullQueue(t *testing.T) {
	processor := &InMemoryProcessor{jobs: make([]Job, 0)}
	queue := CreateQueue(3, 250, processor)

	job1 := &Job{Id: 1}
	result1 := queue.Process(job1)
	assert.Same(t, result1.Job(), job1)
	assert.False(t, result1.Ready())
	assert.Equal(t, "", result1.Content())
	queue.Process(&Job{Id: 2})
	assert.Equal(t, 0, len(processor.jobs))

	job3 := &Job{Id: 3}
	result3 := queue.Process(job3)
	assert.Equal(t, 3, len(processor.jobs))
	assert.True(t, result1.Ready())
	assert.Same(t, result3.Job(), job3)
	assert.True(t, result3.Ready())
	assert.Equal(t, "response: 2", result3.Content())
}
