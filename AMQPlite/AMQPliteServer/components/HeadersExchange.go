package components

import "AMQPlite/AMQPliteServer/frames"

type HeadersExchange struct {
	Name     string
	Type     string
	Bindings map[string]*Binding
}

func (headersExchange *HeadersExchange) Publish(frame frames.FrameEnvelope) {

}

func (headersExchange *HeadersExchange) Delete() {

}

func NewHeadersExchange(name string, exchangeType string) *HeadersExchange {
	return &HeadersExchange{
		Name:     name,
		Type:     exchangeType,
		Bindings: make(map[string]*Binding),
	}
}
