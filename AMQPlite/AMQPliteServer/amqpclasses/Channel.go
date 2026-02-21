package amqpclasses

import "AMQPlite/AMQPliteServer/frames"

type Channel struct {
	ChannelID        uint32
	Inbound          [10]chan frames.GeneralFrame
	Outbound         [10]chan frames.GeneralFrame
	ParentConnection Connection
}
