package amqperrors

import (
	"fmt"
)

// according to AMQP 0-9-1 spec
const (
	// Normal
	ReplySuccess uint16 = 200
	//Channel Exceptions
	ContentTooLarge    uint16 = 311 //The client attempted to transfer content larger than the server could accept at the present time. The client may retry at a later time.
	NoRoute            uint16 = 312
	NoConsumers        uint16 = 313 //When the exchange cannot deliver to a consumer when the immediate flag is set. As a result of pending data on the queue or the absence of any consumers of the queue.
	AccessRefused      uint16 = 403
	NotFound           uint16 = 404
	ResourceLocked     uint16 = 405
	PreconditionFailed uint16 = 406
	//Connection Exceptions
	ConnectionForced uint16 = 320 //An operator intervened to close the connection for some reason. The client may retry at some later date.
	InvalidPath      uint16 = 402
	FrameError       uint16 = 501
	SyntaxError      uint16 = 502
	CommandInvalid   uint16 = 503
	ChannelError     uint16 = 504
	UnexpectedFrame  uint16 = 505
	ResourceError    uint16 = 506
	NotAllowed       uint16 = 530
	NotImplemented   uint16 = 540
	InternalError    uint16 = 541
)

type AMQPError struct {
	Code     uint16
	ClassID  uint16
	MethodID uint16
	Message  string
	IsHard   bool // true for connection exceptions, false for channel
}

func (e *AMQPError) Error() string {
	errType := "channel"
	if e.IsHard {
		errType = "connection"
	}
	return fmt.Sprintf("AMQP %s error %d: %s (class: %d, method: %d)", errType, e.Code, e.Message, e.ClassID, e.MethodID)
}

// close the entire connection
func NewConnectionError(code uint16, classID uint16, methodID uint16, message string) *AMQPError {
	return &AMQPError{
		Code:     code,
		ClassID:  classID,
		MethodID: methodID,
		Message:  message,
		IsHard:   true,
	}
}

// just close the channel
func NewChannelError(code uint16, classID uint16, methodID uint16, message string) *AMQPError {
	return &AMQPError{
		Code:     code,
		ClassID:  classID,
		MethodID: methodID,
		Message:  message,
		IsHard:   false,
	}
}
