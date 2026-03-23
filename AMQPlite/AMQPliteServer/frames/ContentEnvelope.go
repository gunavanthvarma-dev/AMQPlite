package frames

type ContentEnvelope struct {
	RoutingKey string
	Exchange   string
	Header     *ContentHeaderFrame
	Body       []byte
}

func NewContentEnvelope(routingKey string, exchange string, header *ContentHeaderFrame, body []byte) *ContentEnvelope {
	return &ContentEnvelope{
		RoutingKey: routingKey,
		Exchange:   exchange,
		Header:     header,
		Body:       body,
	}
}
