package testqueue

import (
	"sync"
	"time"
)

// Sample of the job. Real-world example would have a payload as well as Id
type Job struct {
	Id int
}

// External dependency handling jobs in batches
type BatchProcessor interface {
	Process([]Job) []string
}

// Queue having ability to aggregate jobs prior to send to BatchProcessor
type Queue interface {
	Process(Job) JobResult
}

type queueImplementation struct {
	maxBatchSize        int
	batchCollectionTime time.Duration
	batchProcessor      BatchProcessor
	jobResults          []*jobResultItem
	flushCancel         chan struct{}
	mutex               sync.Mutex
}

func (queue *queueImplementation) scheduleFlush() {
	for {
		select {
		case <-queue.flushCancel:
			return
		case <-time.After(queue.batchCollectionTime):
			queue.mutex.Lock()
			defer queue.mutex.Unlock()
			if 0 < len(queue.jobResults) {
				go queue.sendBatch(queue.jobResults)
				queue.jobResults = make([]*jobResultItem, 0)
			}

			return
		}
	}
}

func (queue *queueImplementation) sendBatch(resultsBatch []*jobResultItem) {
	batch := make([]Job, 0)

	for _, result := range resultsBatch {
		batch = append(batch, result.job)
	}

	processed := queue.batchProcessor.Process(batch)

	for index, value := range processed {
		resultsBatch[index].setContent(value)
	}
}

func (queue *queueImplementation) Process(job Job) JobResult {
	newJobResult := &jobResultItem{job: job, isReady: make(chan struct{})}

	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.jobResults = append(queue.jobResults, newJobResult)
	batchLength := len(queue.jobResults)

	if queue.maxBatchSize == batchLength {
		close(queue.flushCancel)
		go queue.sendBatch(queue.jobResults)
		queue.jobResults = make([]*jobResultItem, 0)
	} else if batchLength == 1 {
		queue.flushCancel = make(chan struct{})
		go queue.scheduleFlush()
	}

	return newJobResult
}

func NewQueue(maxBatchSize int, batchCollectionTime time.Duration, processor BatchProcessor) Queue {
	return &queueImplementation{
		maxBatchSize:        maxBatchSize,
		batchCollectionTime: batchCollectionTime,
		batchProcessor:      processor,
	}
}
