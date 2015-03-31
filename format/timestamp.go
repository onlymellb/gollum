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
	"github.com/trivago/gollum/core"
	"github.com/trivago/gollum/shared"
)

// Timestamp is a formatter that allows prefixing a message with a timestamp
// (time of arrival at gollum) as well as postfixing it with a delimiter string.
// Configuration example
//
//   - producer.Console
//     Formatter: "format.Timestamp"
//     TimestampDataFormatter: "format.Delimiter"
//     Timestamp: "2006-01-02T15:04:05.000 MST | "
//
// Timestamp defines a Go time format string that is used to format the actual
// timestamp that prefixes the message.
// By default this is set to "2006-01-02 15:04:05 MST | "
//
// TimestampDataFormatter defines the formatter for the data transferred as
// message. By default this is set to "format.Delimiter"
type Timestamp struct {
	core.FormatterBase
	base            core.Formatter
	timestampFormat string
}

func init() {
	shared.RuntimeType.Register(Timestamp{})
}

// Configure initializes this formatter with values from a plugin config.
func (format *Timestamp) Configure(conf core.PluginConfig) error {
	plugin, err := core.NewPluginWithType(conf.GetString("TimestampDataFormatter", "format.Delimiter"), conf)
	if err != nil {
		return err
	}

	format.base = plugin.(core.Formatter)
	format.timestampFormat = conf.GetString("Timestamp", core.DefaultTimestamp)

	return nil
}

// PrepareMessage sets the message to be formatted.
func (format *Timestamp) PrepareMessage(msg core.Message) {
	format.base.PrepareMessage(msg)
	baseLength := format.base.Len()
	timestampStr := msg.Timestamp.Format(format.timestampFormat)

	format.FormatterBase.Message = make([]byte, len(timestampStr)+baseLength)
	len := copy(format.FormatterBase.Message, []byte(timestampStr))
	format.base.Read(format.FormatterBase.Message[len:])
}