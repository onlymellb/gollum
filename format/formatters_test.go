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
	"testing"
)

func testFormatter(t *testing.T, formatter core.Formatter) bool {
	expect := shared.NewExpect(t)

	message := []byte("\ttest\r\n123 456\n")
	msg := core.NewMessage(nil, message, []core.MessageStreamID{}, 0)

	formatter.PrepareMessage(msg)
	buffer := make([]byte, formatter.Len())
	result := true

	length, _ := formatter.Read(buffer)
	result = expect.Equal(formatter.Len(), length) && result
	result = expect.Equal(formatter.Len(), len(formatter.String())) && result
	result = expect.Equal(formatter.Len(), len(formatter.Bytes())) && result
	result = expect.Equal(formatter.String(), string(buffer)) && result

	return result
}

func TestFormatters(t *testing.T) {
	conf := core.PluginConfig{}
	formatters := shared.RuntimeType.GetRegistered("format.")

	if len(formatters) == 0 {
		t.Error("No formatters defined")
	}

	for _, name := range formatters {
		plugin, err := core.NewPluginWithType(name, conf)
		if err != nil {
			t.Errorf("Failed to create formatter %s: %s", name, err.Error())
		} else {
			if !testFormatter(t, plugin.(core.Formatter)) {
				t.Errorf("Formatter %s tests failed", name)
			}
		}
	}
}