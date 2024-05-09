package testqueue

import (
	"time"
)

type Job struct {
	Id int
}

type BatchProcessor interface {
	Process([]Job) []string
}

type Queue interface {
	Process(*Job) JobResult
}

type QueueImplementation struct {
	maxBatchSize        int
	batchCollectionTime time.Duration
	batchProcessor      BatchProcessor
	jobResults          []JobResultItem
	// addingChannel       chan []Job
	// mutex               sync.Mutex
}

// func (queue QueueImplementation) addJob(newJob chan []Job) JobResult {
// 	tick := time.Tick(100 * time.Millisecond)
// 	for {
// 		select {
// 		case <-tick:
// 			fmt.Println("tick.")
// 		case <-newJob:

// 		default:
// 			fmt.Println("    .")
// 			time.Sleep(50 * time.Millisecond)
// 		}
// 	}
// }

func (queue *QueueImplementation) sendIfRequired() {
	if queue.maxBatchSize != len(queue.jobResults) {
		return
	}

	batch := make([]Job, 0)

	for _, result := range queue.jobResults {
		batch = append(batch, *result.job)
	}

	processed := queue.batchProcessor.Process(batch)

	for index, value := range processed {
		queue.jobResults[index].setContent(value)
	}

	queue.jobResults = make([]JobResultItem, 0)
}

func (queue *QueueImplementation) Process(job *Job) JobResult {
	newJobResult := &JobResultItem{job: job, isReady: make(chan struct{}, 1)}
	queue.jobResults = append(queue.jobResults, *newJobResult)

	queue.sendIfRequired()

	return newJobResult
}

func CreateQueue(maxBatchSize int, batchCollectionTime time.Duration, processor BatchProcessor) Queue {
	return &QueueImplementation{
		maxBatchSize:        maxBatchSize,
		batchCollectionTime: batchCollectionTime,
		batchProcessor:      processor,
	}
}
