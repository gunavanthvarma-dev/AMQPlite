package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"errors"
	"log"
)

func ConnWriter(ctx context.Context, connection *Connection) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("Connection is closed")
		case frame := <-connection.WriterChannel:

			switch Frame := frame.(type) {
			case frames.FrameEnvelope:
				log.Println("[DEBUG] ConnWriter: Sending FrameEnvelope")
				frameBytes := Frame.Marshal()
				if _, err := connection.Conn.Write(frameBytes); err != nil {
					return err
				}
			case frames.ContentEnvelope:
				log.Println("[DEBUG] ConnWriter: Sending ContentEnvelope (Header + Body)")
				contentHeaderFrame := Frame.Header
				headerPayload := contentHeaderFrame.Marshal()
				headerFrame := frames.FrameEnvelope{
					FrameType:   2,
					Channel:     contentHeaderFrame.GetChannelID(),
					PayloadSize: uint32(len(headerPayload)),
					Payload:     headerPayload,
				}
				if _, err := connection.Conn.Write(headerFrame.Marshal()); err != nil {
					return err
				}
				bodyFrame := frames.FrameEnvelope{
					FrameType:   3,
					Channel:     Frame.GetChannelID(),
					PayloadSize: uint32(len(Frame.Body)),
					Payload:     Frame.Body,
				}
				if _, err := connection.Conn.Write(bodyFrame.Marshal()); err != nil {
					return err
				}
			}
		}
	}

}
