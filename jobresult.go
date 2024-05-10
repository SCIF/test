package testqueue

import "sync"

type JobResult interface {
	Ready() chan struct{}
	Content() string
	Job() Job
}

type JobResultItem struct {
	job     Job
	content string
	isReady chan struct{}
	mutex   sync.Mutex
}

func (item *JobResultItem) Ready() chan struct{} {
	return item.isReady
}

func (item *JobResultItem) Content() string {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	return item.content
}

func (item *JobResultItem) setContent(response string) {
	item.mutex.Lock()
	defer item.mutex.Unlock()
	item.content = response
	close(item.isReady)
}

func (item *JobResultItem) Job() Job {
	return item.job
}
