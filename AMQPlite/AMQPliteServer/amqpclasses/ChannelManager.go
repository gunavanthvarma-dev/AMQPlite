package amqpclasses

import (
	"AMQPlite/AMQPliteServer/amqperrors"
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"encoding/binary"
)

type ChannelManager struct {
	channels         map[uint16]*Channel
	expectedClassID  uint16
	expectedMethodID uint16
}

func NewChannelManager() ChannelManager {
	return ChannelManager{
		channels: make(map[uint16]*Channel),
	}
}

func (manager *ChannelManager) GetChannel(channelID uint16) (chan<- frames.FrameEnvelope, error) {
	if channel, ok := manager.channels[channelID]; ok {
		return channel.Pipe, nil
	}
	return nil, amqperrors.NewConnectionError(504, 10, 0, "Channel not found")
}

func (manager *ChannelManager) ProcessFrame(InboundChannel chan frames.FrameEnvelope, connection *Connection, ctx context.Context) {
	ctxChan, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
			//print some error
		case frame := <-InboundChannel:

			// get channel id, check if channel exists
			//if its not there, check if its a channel.open frame
			//if its channel.open frame, send it to channel control
			//if channel exists, send it to the channel

			classID := binary.BigEndian.Uint16(frame.Payload[0:2])
			if classID == 20 {
				manager.ChannelControl(frame, connection, ctxChan, cancel)
			} else {
				channel, err := manager.GetChannel(frame.Channel)
				if err != nil {
					// handle error
				} else {
					channel <- frame
				}
			}
		}
	}
}

func (manager *ChannelManager) ChannelControl(frame frames.FrameEnvelope, connection *Connection, ctx context.Context, cancelfunc context.CancelFunc) {

	methodID := binary.BigEndian.Uint16(frame.Payload[2:4])
	switch methodID {
	case 10:
		//channel.open
		manager.channels[frame.Channel] = NewChannel(frame.Channel, connection, ctx, cancelfunc)
		manager.channels[frame.Channel].ParentConnection.WriterChannel <- manager.channels[frame.Channel].SendChannelOpenOK()
	case 20:
		//channel.flow
		//need to implement
	case 21:
		//channel.flow-ok
		//need to implement
	case 40:
		//channel.close
		if ch, exists := manager.channels[frame.Channel]; exists {
			ch.ParentConnection.WriterChannel <- ch.SendChannelCloseOK()
			ch.cancelFunc()
			delete(manager.channels, frame.Channel)
		}
	case 41:
		//channel.close-ok
		//receive channel.close-ok
	}

}
