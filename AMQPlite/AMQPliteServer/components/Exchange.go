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
