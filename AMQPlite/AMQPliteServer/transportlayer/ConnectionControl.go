package transportlayer

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"encoding/binary"
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
			if inboundFrame.FrameType == 8 {
				// Heartbeat, just acknowledge or ignore
				continue
			}
			if inboundFrame.FrameType != 1 {
				// Not a method frame
				continue
			}
			// extract class id and method id
			classID := binary.BigEndian.Uint16(inboundFrame.Payload[0:2])
			methodID := binary.BigEndian.Uint16(inboundFrame.Payload[2:4])
			arguments := inboundFrame.Payload[4:]
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
				//connection.start
				writer <- ConnectionStart()
				connection.SetExpectedClassID(uint16(10))
				connection.SetExpectedMethodID(uint16(11))
			case 11:
				//connection.start-ok
				err := ConnectionStartOK(arguments, connection)
				if err != nil {
					//handle error
				}
				//send connection.tune and wait for connection.tune-ok
				writer <- ConnectionTune(connection)
				connection.SetExpectedClassID(10)
				connection.SetExpectedMethodID(31)
			case 31:
				err := ConnectionTuneOk(arguments, connection)
				connection.ExpectedClassID = 10
				connection.ExpectedMethodID = 40
				if err != nil {
					//handle error
				}
			case 40:
				err := ConnectionOpen(arguments, connection)
				if err != nil {
					//handle error
				}
				writer <- ConnectionOpenOk()
			case 50:
				err := RecvConnectionClose(arguments, connection) //for now the server processes close() received from client, need to find a way to include server closing the connection and waiting for close-ok
				if err != nil {
					//handle error
				}
				// release all resources tied to this connection and then send close-ok
				writer <- ConnectionCloseOk()
			}
			// wrap client message into a frame
			//send it to writer channel

		}
	}
}

func validateFrame(classID uint16, methodID uint16, connection *amqpclasses.Connection) error {
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
