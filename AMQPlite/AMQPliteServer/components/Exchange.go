package components

type Exchange struct {
	Name     string
	Type     string
	Bindings map[string]*Binding
}

type Binding struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

type ExchangeManager struct {
	exchanges map[string]*Exchange
}

func NewExchangeManager() *ExchangeManager {
	return &ExchangeManager{
		exchanges: make(map[string]*Exchange),
	}
}

func (exchangeManager *ExchangeManager) DeclareExchange(name string, exchangeType string) *Exchange {
	exchange := NewExchange(name, exchangeType)
	exchangeManager.exchanges[name] = exchange
	return exchange
}

func (exchangeManager *ExchangeManager) DeleteExchange(name string) {
	delete(exchangeManager.exchanges, name)
}

func (exchangeManager *ExchangeManager) GetExchange(name string) *Exchange {
	return exchangeManager.exchanges[name]
}

func (exchangeManager *ExchangeManager) BindExchange(exchange string, queue string, routingKey string) {
	binding := NewBinding(exchange, queue, routingKey)
	exchangeManager.exchanges[exchange].Bindings[routingKey] = binding
}

func (exchangeManager *ExchangeManager) UnbindExchange(exchange string, queue string, routingKey string) {
	delete(exchangeManager.exchanges[exchange].Bindings, routingKey)
}

func (exchangeManager *ExchangeManager) GetBindings(exchange string) map[string]*Binding {
	return exchangeManager.exchanges[exchange].Bindings
}

func (exchangeManager *ExchangeManager) GetExchanges() map[string]*Exchange {
	return exchangeManager.exchanges
}

func NewExchange(name string, exchangeType string) *Exchange {
	return &Exchange{
		Name:     name,
		Type:     exchangeType,
		Bindings: make(map[string]*Binding),
	}
}

func NewBinding(exchange string, queue string, routingKey string) *Binding {
	return &Binding{
		Exchange:   exchange,
		Queue:      queue,
		RoutingKey: routingKey,
	}
}
