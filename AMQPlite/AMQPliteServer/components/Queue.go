package components

type Queue struct {
	Name string
	Type string
}

func NewQueue(name string, queueType string) *Queue {
	return &Queue{
		Name: name,
		Type: queueType,
	}
}
