package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

type QueueManager struct {
	lock        sync.RWMutex
	queues      map[string]*Queue
	InboundChan chan frames.FrameEnvelope
	broker      *Broker
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues:      make(map[string]*Queue),
		InboundChan: make(chan frames.FrameEnvelope, 10),
	}
}

func (queueManager *QueueManager) SetBroker(broker *Broker) {
	queueManager.broker = broker
}

func (queueManager *QueueManager) GetQueue(queueName string) *Queue {
	queueManager.lock.RLock()
	defer queueManager.lock.RUnlock()
	return queueManager.queues[queueName]
}

func (queueManager *QueueManager) QueueControl(ctx context.Context, writer chan frames.FrameEnvelope) error {
	for {
		select {
		case <-ctx.Done():
			//handle error
			return nil
		case frame := <-queueManager.InboundChan:
			//handle frame
			go func() {
				frame, err := queueManager.handleFrame(frame)
				if err != nil {
					//handle error
					fmt.Println(err)
				}
				writer <- frame
			}()
		}
	}
}

func (queueManager *QueueManager) handleFrame(frame frames.FrameEnvelope) (frames.FrameEnvelope, error) {
	methodID := binary.BigEndian.Uint16(frame.Payload[2:4])
	switch methodID {
	case 10:
		//queue.declare
		queueName := utilties.DecodeShortString(frame.Payload[4:])
		offset := 5 + len(queueName)
		flags := frame.Payload[offset]
		passiveBit := (flags & 0x01) != 0
		durableBit := (flags & 0x02) != 0
		exclusiveBit := (flags & 0x04) != 0
		autoDeleteBit := (flags & 0x08) != 0
		noWaitBit := (flags & 0x10) != 0
		arguments, _, err := utilties.DecodeFieldTable(frame.Payload[offset+1:])
		if err != nil {
			fmt.Println(err)
			return frames.FrameEnvelope{}, err
		}
		queueManager.DeclareQueue(queueName, passiveBit, durableBit, exclusiveBit, autoDeleteBit, noWaitBit, arguments)
		return queueManager.DeclareQueueOK(queueName)
	case 20:
		//queue.bind
		//reserved1 := binary.BigEndian.Uint16(frame.Payload[4:6])
		queueName := utilties.DecodeShortString(frame.Payload[6:])
		exchangeName := utilties.DecodeShortString(frame.Payload[7+len(queueName):])
		routingKey := utilties.DecodeShortString(frame.Payload[8+len(queueName)+len(exchangeName):])
		//noWaitBit := (frame.Payload[9+len(queueName)+len(exchangeName)+len(routingKey)] & 0x01) != 0
		arguments, _, err := utilties.DecodeFieldTable(frame.Payload[10+len(queueName)+len(exchangeName)+len(routingKey):])
		if err != nil {
			fmt.Println(err)
			return frames.FrameEnvelope{}, err
		}
		queueManager.BindQueue(queueName, exchangeName, routingKey, arguments)
		return queueManager.BindQueueOK(queueName)
	case 40:
		//queue.delete
		//reserved1 := binary.BigEndian.Uint16(frame.Payload[4:6])
		queueName := utilties.DecodeShortString(frame.Payload[6:])
		offset := 7 + len(queueName)
		flags := frame.Payload[offset]
		ifUnusedBit := (flags & 0x01) != 0
		ifEmptyBit := (flags & 0x02) != 0
		noWaitBit := (flags & 0x04) != 0
		arguments, _, err := utilties.DecodeFieldTable(frame.Payload[offset+1:])
		if err != nil {
			fmt.Println(err)
			return frames.FrameEnvelope{}, err
		}
		queueManager.DeleteQueue(queueName, ifUnusedBit, ifEmptyBit, noWaitBit, arguments)
		return queueManager.DeleteQueueOK(queueName)

	default:
		return frames.FrameEnvelope{}, errors.New("unknown method id")
	}
}

func (queueManager *QueueManager) DeclareQueue(queueName string, passiveBit bool, durableBit bool, exclusiveBit bool, autoDeleteBit bool, noWaitBit bool, arguments map[string]any) error {
	queueManager.lock.Lock()
	defer queueManager.lock.Unlock()
	if _, ok := queueManager.queues[queueName]; ok {
		return nil
	}
	queueManager.queues[queueName] = NewQueue(queueName, durableBit, exclusiveBit, autoDeleteBit)
	return nil
}

func (queueManager *QueueManager) DeclareQueueOK(queueName string) (frames.FrameEnvelope, error) {
	queue := queueManager.queues[queueName]
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(queue.Name))
	binary.Write(payloadbuf, binary.BigEndian, queue.MessageCount)
	binary.Write(payloadbuf, binary.BigEndian, queue.ConsumerCount)
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}

func (queueManager *QueueManager) BindQueue(queueName string, exchangeName string, routingKey string, arguments map[string]any) error {
	queueManager.lock.Lock()
	defer queueManager.lock.Unlock()
	queue := queueManager.queues[queueName]
	if queue == nil {
		return errors.New("queue not found")
	}
	binding := NewBinding(exchangeName, queueName, routingKey) //create binding
	if _, ok := queue.ExchangeBindings[exchangeName]; !ok {
		queue.ExchangeBindings[exchangeName] = make([]*Binding, 0)
	}
	for _, b := range queue.ExchangeBindings[exchangeName] {
		if b.RoutingKey == routingKey {
			return nil
		}
	}
	queue.ExchangeBindings[exchangeName] = append(queue.ExchangeBindings[exchangeName], binding) //add binding to queue
	exchange, err := queueManager.broker.ExchangeManager.GetExchange(exchangeName)
	if err != nil {
		return err
	}
	exchange.AddBinding(binding, queueName)
	return nil
}

func (queueManager *QueueManager) BindQueueOK(queueName string) (frames.FrameEnvelope, error) {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(queueName))
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}

func (queueManager *QueueManager) UnbindQueue(queueName string, exchangeName string, routingKey string) error {
	queueManager.lock.Lock()
	defer queueManager.lock.Unlock()
	queue := queueManager.queues[queueName]
	if queue == nil {
		return errors.New("queue not found")
	}
	for i, b := range queue.ExchangeBindings[exchangeName] {
		if b.RoutingKey == routingKey {
			queue.ExchangeBindings[exchangeName] = append(queue.ExchangeBindings[exchangeName][:i], queue.ExchangeBindings[exchangeName][i+1:]...)
			return nil
		}
	}
	return nil
}

func (queueManager *QueueManager) UnbindQueueOK(queueName string) (frames.FrameEnvelope, error) {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(queueName))
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}

func (queueManager *QueueManager) PurgeQueue(queueName string, noWaitBit bool) error {
	queueManager.lock.Lock()
	defer queueManager.lock.Unlock()
	queue := queueManager.queues[queueName]
	if queue == nil {
		return errors.New("queue not found")
	}
	queue.MessageCount = 0
	return nil
}

func (queueManager *QueueManager) PurgeQueueOK(queueName string) (frames.FrameEnvelope, error) {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(queueName))
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}

func (queueManager *QueueManager) DeleteQueue(queueName string, ifUnusedBit bool, ifEmptyBit bool, noWaitBit bool, arguments map[string]any) error {
	queueManager.lock.Lock()
	defer queueManager.lock.Unlock()
	if _, ok := queueManager.queues[queueName]; !ok {
		return nil
	}
	delete(queueManager.queues, queueName)
	return nil
}

func (queueManager *QueueManager) DeleteQueueOK(queueName string) (frames.FrameEnvelope, error) {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, queueManager.queues[queueName].MessageCount)
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}
