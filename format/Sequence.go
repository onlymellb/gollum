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

package format

import (
	"fmt"
	"github.com/trivago/gollum/core"
	"github.com/trivago/gollum/shared"
	"io"
)

// Sequence is a formatter that allows prefixing a message with the message's
// sequence number
// Configuration example
//
//   - producer.Console
//     Formatter: "format.Sequence"
//     SequenceDataFormatter: "format.Delimiter"
//
// SequenceDataFormatter defines the formatter for the data transferred as
// message. By default this is set to "format.Forward"
type Sequence struct {
	base     core.Formatter
	length   int
	sequence string
}

func init() {
	shared.RuntimeType.Register(Sequence{})
}

// Configure initializes this formatter with values from a plugin config.
func (format *Sequence) Configure(conf core.PluginConfig) error {
	plugin, err := core.NewPluginWithType(conf.GetString("SequenceDataFormatter", "format.Forward"), conf)
	if err != nil {
		return err
	}

	format.base = plugin.(core.Formatter)
	return nil
}

// PrepareMessage sets the message to be formatted.
func (format *Sequence) PrepareMessage(msg core.Message) {
	format.base.PrepareMessage(msg)
	format.sequence = fmt.Sprintf("%d:", msg.Sequence)
	format.length = format.base.Len() + len(format.sequence)
}

// Len returns the length of a formatted message.
func (format *Sequence) Len() int {
	return format.length
}

// String returns the message as string
func (format *Sequence) String() string {
	return fmt.Sprintf("%s%s", format.sequence, format.base.String())
}

// CopyTo copies the message into an existing buffer. It is assumed that
// dest has enough space to fit GetLength() bytes
func (format *Sequence) Read(dest []byte) (int, error) {
	len := copy(dest, []byte(format.sequence))
	baseLen, err := format.base.Read(dest[len:])
	return len + baseLen, err
}

// WriteTo implements the io.WriterTo interface.
// Data will be written directly to a writer.
func (format *Sequence) WriteTo(writer io.Writer) (int64, error) {
	len, err := writer.Write([]byte(format.sequence))
	if err != nil {
		return int64(len), err
	}

	var baseLen int64
	baseLen, err = format.base.WriteTo(writer)
	return int64(len) + baseLen, err
}
