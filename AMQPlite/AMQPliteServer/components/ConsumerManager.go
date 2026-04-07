package components

import (
	"errors"
	"sync"
)

type ConsumerManager struct {
	Consumers map[string]string
	lock      sync.RWMutex
}

func NewConsumerManager() *ConsumerManager {
	return &ConsumerManager{Consumers: make(map[string]string)}
}

func (manager *ConsumerManager) GetQueueAttachedToConsumer(consumerTag string) (string, error) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	queue, ok := manager.Consumers[consumerTag]
	if ok == true {
		return queue, nil
	} else {
		return "", errors.New("No queue attached to the given consumer or no consumer exists")
	}
}

func (manager *ConsumerManager) AddConsumer(consumerTag string, queue string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	_, ok := manager.Consumers[consumerTag]
	if ok == false {
		manager.Consumers[consumerTag] = queue
		return nil
	} else {
		return errors.New("consumerTag already exists")
	}
}

func (manager *ConsumerManager) RemoveConsumer(consumerTag string) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	delete(manager.Consumers, consumerTag)
}
