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
	InboundChan chan frames.ChannelEnvelope
	broker      *Broker
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewQueueManager(ctx context.Context, cancel context.CancelFunc) *QueueManager {
	return &QueueManager{
		queues:      make(map[string]*Queue),
		InboundChan: make(chan frames.ChannelEnvelope, 10),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (queueManager *QueueManager) SetBroker(broker *Broker) {
	queueManager.broker = broker
}

func (queueManager *QueueManager) GetQueue(queueName string) (*Queue, error) {
	queueManager.lock.RLock()
	defer queueManager.lock.RUnlock()
	queue, ok := queueManager.queues[queueName]
	if !ok {
		return nil, errors.New("queue not found")
	}
	return queue, nil
}

func (queueManager *QueueManager) QueueControl() error {
	for {
		select {
		case <-queueManager.ctx.Done():
			//handle error
			return nil
		case frame := <-queueManager.InboundChan:
			//handle frame
			outputFrame, err := queueManager.handleFrame(frame.Frame)
			if err != nil {
				//handle error
				fmt.Println(err)
			}
			frame.ChannelCallback <- outputFrame
		}
	}
}

func (queueManager *QueueManager) handleFrame(frame frames.FrameEnvelope) (frames.FrameEnvelope, error) {
	methodID := binary.BigEndian.Uint16(frame.Payload[2:4])
	switch methodID {
	case 10:
		//queue.declare
		//reserved_1 = frame.Payload[4:6]
		queueName := utilties.DecodeShortString(frame.Payload[6:])
		offset := 7 + len(queueName)
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
		queueManager.BindQueue(queueName, "", queueName, nil) // Bind without creating a deadlock context
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
	queueCtx, queueCancel := context.WithCancel(queueManager.ctx)
	newQueue := NewQueue(queueName, durableBit, exclusiveBit, autoDeleteBit, queueCtx, queueCancel)
	queueManager.queues[queueName] = newQueue
	go newQueue.ForwardMessages() // Start queue listener
	return nil
}

func (queueManager *QueueManager) DeclareQueueOK(queueName string) (frames.FrameEnvelope, error) {
	queue := queueManager.queues[queueName]
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(50)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(11)) // Method ID: DeclareOk
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
	binary.Write(payloadbuf, binary.BigEndian, uint16(50)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(21)) // Method ID: BindOk
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
	binary.Write(payloadbuf, binary.BigEndian, uint16(50)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(51)) // Method ID: UnbindOk
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
	binary.Write(payloadbuf, binary.BigEndian, uint16(50)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(31)) // Method ID: PurgeOk
	// PurgeOk replies with message-count
	if q, ok := queueManager.queues[queueName]; ok {
		binary.Write(payloadbuf, binary.BigEndian, q.MessageCount)
	} else {
		binary.Write(payloadbuf, binary.BigEndian, uint32(0))
	}
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}

func (queueManager *QueueManager) DeleteQueue(queueName string, ifUnusedBit bool, ifEmptyBit bool, noWaitBit bool, arguments map[string]any) error {
	queueManager.lock.Lock()
	defer queueManager.lock.Unlock()
	if queue, ok := queueManager.queues[queueName]; ok {
		if ifUnusedBit && queue.MessageCount > 0 {
			return errors.New("queue is not empty")
		}
		if ifEmptyBit && queue.MessageCount > 0 {
			return errors.New("queue is not empty")
		}
		queue.cancel()
		delete(queueManager.queues, queueName)
	}
	return nil
}

func (queueManager *QueueManager) DeleteQueueOK(queueName string) (frames.FrameEnvelope, error) {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(50)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(41)) // Method ID: DeleteOk
	// DeleteOk replies with message-count (0 since it's deleted)
	binary.Write(payloadbuf, binary.BigEndian, uint32(0))
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame, nil
}
