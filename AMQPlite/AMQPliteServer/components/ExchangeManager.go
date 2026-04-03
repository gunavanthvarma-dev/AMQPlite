package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

type ExchangeManager struct {
	lock         sync.RWMutex
	exchanges    map[string]Exchange
	InboundChan  chan frames.ChannelEnvelope
	OutboundChan chan frames.FrameEnvelope
	ctx          context.Context
	cancel       context.CancelFunc
	broker       *Broker
}

func NewExchangeManager(ctx context.Context, cancel context.CancelFunc) *ExchangeManager {
	return &ExchangeManager{
		exchanges:    make(map[string]Exchange),
		InboundChan:  make(chan frames.ChannelEnvelope, 10),
		OutboundChan: make(chan frames.FrameEnvelope, 10),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (exchangeManager *ExchangeManager) SetBroker(broker *Broker) {
	exchangeManager.broker = broker
}

func (exchangeManager *ExchangeManager) ExchangeControl() error {
	for {
		select {
		case <-exchangeManager.ctx.Done():
			//handle error
			return nil
		case frame := <-exchangeManager.InboundChan:
			//handle frame

			outputFrame, err := exchangeManager.handleFrame(frame.Frame)
			if err != nil {
				//handle error
			}
			frame.ChannelCallback <- outputFrame
		}
	}
}

func (exchangeManager *ExchangeManager) handleFrame(frame frames.FrameEnvelope) (frames.FrameEnvelope, error) {
	methodID := binary.BigEndian.Uint16(frame.Payload[2:4])
	switch methodID {
	case 10:
		//exchange.declare
		//reserved_1 := frame.Payload[3:5]
		exchangeNameLength := binary.BigEndian.Uint16(frame.Payload[5:7])
		exchangeName := string(frame.Payload[7 : 7+exchangeNameLength])
		exchangeTypeLength := binary.BigEndian.Uint16(frame.Payload[7+exchangeNameLength : 7+exchangeNameLength+2])
		exchangeType := string(frame.Payload[7+exchangeNameLength+2 : 7+exchangeNameLength+2+exchangeTypeLength])
		//passiveBit := frame.Payload[7+exchangeNameLength+2+exchangeTypeLength]
		//durableBit := frame.Payload[7+exchangeNameLength+2+exchangeTypeLength+1]
		//reserved_2 := frame.Payload[7+exchangeNameLength+2+exchangeTypeLength+2]
		//reserved_3 := frame.Payload[7+exchangeNameLength+2+exchangeTypeLength+3]
		//no_waitBit := frame.Payload[7+exchangeNameLength+2+exchangeTypeLength+4]
		//argumentsLength := binary.BigEndian.Uint16(frame.Payload[7+exchangeNameLength+2+exchangeTypeLength+5 : 7+exchangeNameLength+2+exchangeTypeLength+7])
		//arguments := frame.Payload[7+exchangeNameLength+2+exchangeTypeLength+7 : 7+exchangeNameLength+2+exchangeTypeLength+7+argumentsLength]
		exchangeContext, cancel := context.WithCancel(exchangeManager.ctx)
		_, err := exchangeManager.DeclareExchange(exchangeName, exchangeType, exchangeContext, cancel)
		if err != nil {
			fmt.Println(err)
			return frames.FrameEnvelope{}, err
		}
		return exchangeManager.ExchangeDeclareOk(), nil
	case 20:
		//exchange.delete
		exchangeNameLength := binary.BigEndian.Uint16(frame.Payload[5:7])
		exchangeName := string(frame.Payload[7 : 7+exchangeNameLength])
		//no_waitBit := frame.Payload[7+exchangeNameLength]
		err := exchangeManager.DeleteExchange(exchangeName)
		if err != nil {
			fmt.Println(err)
			return frames.FrameEnvelope{}, err
		}
		return exchangeManager.ExchangeDeleteOk(), nil
	default:
		return frames.FrameEnvelope{}, errors.New("unknown method id")
	}

}

func (exchangeManager *ExchangeManager) DeclareExchange(name string, exchangeType string, ctx context.Context, cancel context.CancelFunc) (Exchange, error) {
	exchange, ok := exchangeManager.exchanges[name]
	if ok {
		return nil, errors.New("exchange already exists")
	}
	exchange = NewExchange(name, exchangeType, exchangeManager, ctx, cancel)
	exchangeManager.exchanges[name] = exchange
	go exchange.Publish() //launching a goroutine for each exchange to handle the incoming messages
	return exchange, nil
}

func (exchangeManager *ExchangeManager) DeleteExchange(name string) error {
	exchange, ok := exchangeManager.exchanges[name]
	if !ok {
		return errors.New("exchange not found")
	}
	exchange.Cancel()
	delete(exchangeManager.exchanges, name)
	return nil
}

func (exchangeManager *ExchangeManager) GetExchange(name string) (Exchange, error) {
	exchange, ok := exchangeManager.exchanges[name]
	if !ok {
		return nil, errors.New("exchange not found")
	}
	return exchange, nil
}

// func (exchangeManager *ExchangeManager) BindExchange(exchange string, queue string, routingKey string) {
// 	binding := NewBinding(exchange, queue, routingKey)
// 	exchangeManager.exchanges[exchange].Bindings[routingKey] = binding
// }

// func (exchangeManager *ExchangeManager) UnbindExchange(exchange string, queue string, routingKey string) {
// 	delete(exchangeManager.exchanges[exchange].Bindings, routingKey)
// }

// func (exchangeManager *ExchangeManager) GetBindings(exchange string) map[string]*Binding {
// 	return exchangeManager.exchanges[exchange].Bindings
// }

func (exchangeManager *ExchangeManager) GetExchanges() map[string]Exchange {
	return exchangeManager.exchanges
}

func (exchangeManager *ExchangeManager) ExchangeDeclareOk() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	frame.FrameType = 1
	frame.Channel = 0
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.BigEndian, uint16(40))
	binary.Write(payload, binary.BigEndian, uint16(11))
	frame.PayloadSize = uint32(len(payload.Bytes()))
	frame.Payload = payload.Bytes()
	return frame
}

func (exchangeManager *ExchangeManager) ExchangeDeleteOk() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	frame.FrameType = 1
	frame.Channel = 0
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.BigEndian, uint16(40))
	binary.Write(payload, binary.BigEndian, uint16(21))
	frame.PayloadSize = uint32(len(payload.Bytes()))
	frame.Payload = payload.Bytes()
	return frame
}
