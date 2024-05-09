package testqueue

type JobResult interface {
	Ready() bool
	Content() string
	Job() *Job
}

type JobResultItem struct {
	job         *Job
	content     string
	isReady     chan bool
	isReadyBool bool
}

func (item *JobResultItem) Ready() bool {
	_, notClosed := (<-item.isReady)

	return !notClosed
}

func (item JobResultItem) Content() string {
	return item.content
}

func (item JobResultItem) Job() *Job {
	return item.job
}
