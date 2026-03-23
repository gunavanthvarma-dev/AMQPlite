package frames

type ContentEnvelope struct {
	RoutingKey string
	Exchange   string
	Header     *ContentHeaderFrame
	Body       []byte
}

func NewContentEnvelope(exchange string, routingKey string, header *ContentHeaderFrame, body []byte) *ContentEnvelope {
	return &ContentEnvelope{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Header:     header,
		Body:       body,
	}
}
