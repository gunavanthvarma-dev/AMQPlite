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
	InboundChan  chan frames.FrameEnvelope
	OutboundChan chan frames.FrameEnvelope
	broker       *Broker
}

func NewExchangeManager() *ExchangeManager {
	return &ExchangeManager{
		exchanges:    make(map[string]Exchange),
		InboundChan:  make(chan frames.FrameEnvelope, 10),
		OutboundChan: make(chan frames.FrameEnvelope, 10),
	}
}

func (exchangeManager *ExchangeManager) SetBroker(broker *Broker) {
	exchangeManager.broker = broker
}

func (exchangeManager *ExchangeManager) ExchangeControl(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			//handle error
			return nil
		case frame := <-exchangeManager.InboundChan:
			//handle frame
			err := exchangeManager.handleFrame(frame)
			if err != nil {
				//handle error
			}
		}
	}
}

func (exchangeManager *ExchangeManager) handleFrame(frame frames.FrameEnvelope) error {
	methodID := binary.BigEndian.Uint16(frame.Payload[1:3])
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
		_, err := exchangeManager.DeclareExchange(exchangeName, exchangeType)
		if err != nil {
			fmt.Println(err)
			return err
		}
		exchangeManager.OutboundChan <- exchangeManager.ExchangeDeclareOk()
		return nil
	case 20:
		//exchange.delete
		exchangeNameLength := binary.BigEndian.Uint16(frame.Payload[5:7])
		exchangeName := string(frame.Payload[7 : 7+exchangeNameLength])
		//no_waitBit := frame.Payload[7+exchangeNameLength]
		err := exchangeManager.DeleteExchange(exchangeName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		exchangeManager.OutboundChan <- exchangeManager.ExchangeDeleteOk()
		return nil
	default:
		return errors.New("unknown method id")
	}

}

func (exchangeManager *ExchangeManager) DeclareExchange(name string, exchangeType string) (Exchange, error) {
	exchange, ok := exchangeManager.exchanges[name]
	if ok {
		return nil, errors.New("exchange already exists")
	}
	exchange = NewExchange(name, exchangeType)
	exchangeManager.exchanges[name] = exchange
	return exchange, nil
}

func (exchangeManager *ExchangeManager) DeleteExchange(name string) error {
	_, ok := exchangeManager.exchanges[name]
	if !ok {
		return errors.New("exchange not found")
	}
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
