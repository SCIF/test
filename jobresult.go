package testqueue

import "sync"

// The result object. Ready() method allow to wait for the result resolving or just check the current status.
// The copy of original Job is stored in order to simplify matching job and result.
// Real-world example most likely would use typed payload rather than just string of the response
type JobResult interface {
	Ready() chan struct{}
	Content() string
	Job() Job
}

type jobResultItem struct {
	job     Job
	content string
	isReady chan struct{}
	mutex   sync.Mutex
}

func (item *jobResultItem) Ready() chan struct{} {
	return item.isReady
}

func (item *jobResultItem) Content() string {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	return item.content
}

func (item *jobResultItem) setContent(response string) {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	item.content = response
	close(item.isReady)
}

func (item *jobResultItem) Job() Job {
	return item.job
}
