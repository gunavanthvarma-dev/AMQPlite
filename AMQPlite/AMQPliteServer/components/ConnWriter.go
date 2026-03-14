package components

import (
	"context"
	"errors"
)

func ConnWriter(ctx context.Context, connection *Connection) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("Connection is closed")
		case frame := <-connection.WriterChannel:
			frameBytes := frame.Marshal()
			if _, err := connection.Conn.Write(frameBytes); err != nil {
				return err
			}
		}
	}

}
