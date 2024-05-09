package testqueue

type JobResult interface {
	Ready() bool
	Content() string
	Job() *Job
}

type JobResultItem struct {
	job     *Job
	content string
	isReady chan struct{}
}

func (item *JobResultItem) Ready() bool {
	select {
	case <-item.isReady:
		_, isOpen := <-item.isReady
		return !isOpen
	default:
		return false
	}
}

func (item *JobResultItem) Content() string {
	return item.content
}

func (item *JobResultItem) setContent(response string) {
	item.content = response
	close(item.isReady)
}

// debug
func (item *JobResultItem) Job() *Job {
	return item.job
}
