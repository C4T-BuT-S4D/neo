package queue

type simpleQueueFactory struct{}

func (f simpleQueueFactory) Create(maxJobs int) Queue {
	return NewSimpleQueue(maxJobs)
}

func (f simpleQueueFactory) Type() Type {
	return TypeSimple
}

func NewSimpleQueueFactory() Factory {
	return simpleQueueFactory{}
}
