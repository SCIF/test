package testqueue

import (
	"sync"
	"time"
)

type Job struct {
	Id int
}

type BatchProcessor interface {
	Process([]Job) []string
}

type Queue interface {
	Process(Job) JobResult
}

type QueueImplementation struct {
	maxBatchSize        int
	batchCollectionTime time.Duration
	batchProcessor      BatchProcessor
	jobResults          []*JobResultItem
	flushCancel         chan struct{}
	mutex               sync.Mutex
}

func (queue *QueueImplementation) scheduleFlush() {
	for {
		select {
		case <-queue.flushCancel:
			return
		case <-time.After(queue.batchCollectionTime):
			queue.mutex.Lock()
			defer queue.mutex.Unlock()
			if 0 < len(queue.jobResults) {
				go queue.sendBatch(queue.jobResults)
				queue.jobResults = make([]*JobResultItem, 0)
			}

			return
		}
	}
}

func (queue *QueueImplementation) sendBatch(resultsBatch []*JobResultItem) {
	batch := make([]Job, 0)

	for _, result := range resultsBatch {
		batch = append(batch, result.job)
	}

	processed := queue.batchProcessor.Process(batch)

	for index, value := range processed {
		resultsBatch[index].setContent(value)
	}
}

func (queue *QueueImplementation) Process(job Job) JobResult {
	newJobResult := &JobResultItem{job: job, isReady: make(chan struct{})}

	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.jobResults = append(queue.jobResults, newJobResult)
	batchLength := len(queue.jobResults)

	if queue.maxBatchSize == batchLength {
		close(queue.flushCancel)
		go queue.sendBatch(queue.jobResults)
		queue.jobResults = make([]*JobResultItem, 0)
	} else if batchLength == 1 {
		queue.flushCancel = make(chan struct{})
		go queue.scheduleFlush()
	}

	return newJobResult
}

func CreateQueue(maxBatchSize int, batchCollectionTime time.Duration, processor BatchProcessor) Queue {
	return &QueueImplementation{
		maxBatchSize:        maxBatchSize,
		batchCollectionTime: batchCollectionTime,
		batchProcessor:      processor,
	}
}
