package queue

type endlessQueueFactory struct{}

func (f endlessQueueFactory) Create(maxJobs int) Queue {
	return NewEndlessQueue(maxJobs)
}

func (f endlessQueueFactory) Type() Type {
	return TypeEndless
}

func NewEndlessQueueFactory() Factory {
	return endlessQueueFactory{}
}
