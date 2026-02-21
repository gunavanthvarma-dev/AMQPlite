package amqpclasses

type ChannelManager struct {
	channels map[int32]Channel
}

func NewChannelManager() ChannelManager {
	return ChannelManager{
		channels: make(map[int32]Channel),
	}
}
