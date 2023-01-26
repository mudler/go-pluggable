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
	"fmt"
)

// EventType describes an event type
type EventType string

// Event describes the event structure.
// Contains a Name field and a Data field which
// is marshalled in JSON
type Event struct {
	Name EventType `json:"name"`
	Data string    `json:"data"`
	File string    `json:"file"` // If Data >> 10K write content to file instead
}

// EventResponse describes the event response structure
// It represent the JSON response from plugins
type EventResponse struct {
	State string `json:"state"`
	Data  string `json:"data"`
	Error string `json:"error"`
	Logs  string `json:"log"`
}

// JSON returns the stringified JSON of the Event
func (e Event) JSON() (string, error) {
	dat, err := json.Marshal(e)
	return string(dat), err
}

// Copy returns a copy of Event
func (e Event) Copy() *Event {
	copy := &e
	return copy
}

func (e Event) ResponseEventName(s string) EventType {
	return EventType(fmt.Sprintf("%s-%s", e.Name, s))
}

// Unmarshal decodes the json payload in the given parameteer
func (r EventResponse) Unmarshal(i interface{}) error {
	return json.Unmarshal([]byte(r.Data), i)
}

// Errored returns true if the response contains an error
func (r EventResponse) Errored() bool {
	return len(r.Error) != 0
}

// NewEvent returns a new event which can be used for publishing
// the obj gets automatically serialized in json.
func NewEvent(name EventType, obj interface{}) (*Event, error) {
	dat, err := json.Marshal(obj)
	return &Event{Name: name, Data: string(dat)}, err
}
