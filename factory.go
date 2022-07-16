/*
Copyright © 2020-2022 Ettore Di Giacinto <mudler@mocaccino.org>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pluggable

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type FactoryPlugin struct {
	EventType     EventType
	PluginHandler PluginHandler
}

func NewPluginFactory(p ...FactoryPlugin) PluginFactory {
	f := make(PluginFactory)
	for _, pp := range p {
		f.Add(pp.EventType, pp.PluginHandler)
	}
	return f
}

// PluginHandler represent a generic plugin which
// talks go-pluggable API
// It receives an event, and is always expected to give a response
type PluginHandler func(*Event) EventResponse

// PluginFactory is a collection of handlers for a given event type.
// a plugin has to respond to multiple events and it always needs to return an
// Event response as result
type PluginFactory map[EventType]PluginHandler

// Run runs the PluginHandler given a event type and a payload
//
// The result is written to the writer provided
// as argument.
func (p PluginFactory) Run(name EventType, r io.Reader, w io.Writer) error {
	ev := &Event{}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, ev); err != nil {
		return err
	}

	if ev.File != "" {
		c, err := ioutil.ReadFile(ev.File)
		if err != nil {
			return err
		}

		ev.Data = string(c)
	}

	resp := EventResponse{}
	for e, r := range p {
		if name == e {
			resp = r(ev)
		}
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(dat)
	return err
}

// Add associates an handler to an event type
func (p PluginFactory) Add(ev EventType, ph PluginHandler) {
	p[ev] = ph
}
