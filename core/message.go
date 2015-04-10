// Copyright 2015 trivago GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"runtime"
	"time"
)

// MessageStreamID is the "compiled name" of a stream
type MessageStreamID uint64

const (
	// LogInternalStream is the name of the internal message channel (logs)
	LogInternalStream = "_GOLLUM_"
	// WildcardStream is the name of the "all streams" channel
	WildcardStream = "*"
	// DroppedStream is the name of the stream used to store dropped messages
	DroppedStream = "_DROPPED_"
)

var (
	// LogInternalStreamID is the ID of the "_GOLLUM_" stream
	LogInternalStreamID = GetStreamID(LogInternalStream)

	// WildcardStreamID is the ID of the "*" stream
	WildcardStreamID = GetStreamID(WildcardStream)

	// DroppedStreamID is the ID of the "_DROPPED_" stream
	DroppedStreamID = GetStreamID(DroppedStream)
)

var messageRetryQueue chan Message

// MessageSource defines methods that can be called on types that generate
// messages.
type MessageSource interface {
	// Pause instructs the source to stop sending messages.
	Pause()

	// IsPaused returns true if the source is currently in the paused state.
	IsPaused() bool

	// Resume instructs the source to start sending messages again.
	Resume()
}

// Message is a container used for storing the internal state of messages.
// This struct is passed between consumers and producers.
type Message struct {
	Data      []byte
	StreamID  MessageStreamID
	Source    MessageSource
	Timestamp time.Time
	Sequence  uint64
}

// EnableRetryQueue creates a retried messages channel using the given size.
func EnableRetryQueue(size int) {
	if messageRetryQueue == nil {
		messageRetryQueue = make(chan Message, size)
	}
}

// GetRetryQueue returns read access to the retry queue.
func GetRetryQueue() <-chan Message {
	return messageRetryQueue
}

// NewMessage creates a new message from a given data stream
func NewMessage(source MessageSource, data []byte, sequence uint64) Message {
	return Message{
		Data:      data,
		Source:    source,
		StreamID:  WildcardStreamID,
		Timestamp: time.Now(),
		Sequence:  sequence,
	}
}

// Enqueue is a convenience function to push a message to a channel while
// waiting for a timeout instead of just blocking.
// Passing a timeout of -1 will discard the message.
// Passing a timout of 0 will always block.
// Messages that time out will be passed to the dropped queue if a Dropped
// consumer exists.
func (msg Message) Enqueue(channel chan<- Message, timeout time.Duration) {
	if timeout == 0 {
		channel <- msg
		return // ### return, done ###
	}

	start := time.Time{}
	for {
		select {
		case channel <- msg:
			return // ### return, done ###

		default:
			switch {
			// Start timeout based retries
			case start.IsZero():
				if timeout < 0 {
					return // ### return, drop and ignore ###
				}
				start = time.Now()

			// Discard message after timeout
			case time.Since(start) > timeout:
				go msg.Drop(time.Duration(0))
				return // ### return, drop and retry ###

			// Yield and try again
			default:
				runtime.Gosched()
			}
		}
	}
}

// String implements the stringer interface
func (msg Message) String() string {
	return string(msg.Data)
}

// Retry pushes a message to the retry queue. This queue can be consumed by the
// loopback consumer. If no such consumer has been configured, the message is
// lost.
func (msg Message) Retry(streamID MessageStreamID, timeout time.Duration) {
	if messageRetryQueue != nil {
		msg.StreamID = streamID
		msg.Enqueue(messageRetryQueue, timeout)
	}
}

// Drop is a shortcut for msg.Retry(DroppedStreamID, timeout)
func (msg Message) Drop(timeout time.Duration) {
	msg.Retry(DroppedStreamID, timeout)
}
