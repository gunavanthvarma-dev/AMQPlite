package transportlayer

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"log"
)

func ConnectionControl(Inbound chan frames.FrameEnvelope, writer chan frames.GeneralFrame, ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			//send error to client
			log.Fatalf("closing connection")
		case inboundFrame := <-Inbound:
			//handle inboundFrame by implementing Connection class
			// wrap client message into a frame
			//send it to writer channel
		}
	}
}
