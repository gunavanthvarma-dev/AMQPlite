package components

type Binding struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

func NewBinding(exchange string, queue string, routingKey string) *Binding {
	return &Binding{
		Exchange:   exchange,
		Queue:      queue,
		RoutingKey: routingKey,
	}
}
