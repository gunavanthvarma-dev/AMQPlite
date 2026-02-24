package transportlayer

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"errors"
	"fmt"
	"log"
)

func ConnectionControl(Inbound chan frames.FrameEnvelope, writer chan frames.FrameEnvelope, ctx context.Context, connection *amqpclasses.Connection) {

	for {
		select {
		case <-ctx.Done():
			//send error to client
			log.Fatalf("closing connection")
		case inboundFrame := <-Inbound:
			// extract class id and method id
			classID := uint8(inboundFrame.Payload[0])
			methodID := uint8(inboundFrame.Payload[1])
			// call frame validation function
			err := validateFrame(classID, methodID, connection)
			if err != nil {
				//handle error
			}
			// --- If the frame is valid, handle it
			// --- else throw appropriate error
			//handle inboundFrame by implementing Connection class
			switch methodID {
			case 10:
				writer <- ConnectionStart()
			}
			// wrap client message into a frame
			//send it to writer channel

		}
	}
}

func validateFrame(classID uint8, methodID uint8, connection *amqpclasses.Connection) error {
	//lock RW
	connection.Lock.RLock()
	defer connection.Lock.RUnlock()

	if connection.ExpectedClassID != classID {
		return errors.New("Incorrect class, Connection class required!!!")
	}
	if connection.ExpectedMethodID != 0 && connection.ExpectedMethodID != methodID {
		return fmt.Errorf("Unexpected frame from client: Expected Method Id:%d Received:%d", connection.ExpectedMethodID, methodID)
	}
	//check if the expectedMethodID
	return nil
}
