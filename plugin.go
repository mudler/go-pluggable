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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// Plugin describes binaries to be hooked on events, with common js input, and common js output
type Plugin struct {
	Name       string
	Executable string
}

// A safe threshold to avoid unpleasant exec buffer fill for argv too big. Seems 128K is the limit on Linux.
const maxMessageSize = 1 << 13

// Run runs the Event on the plugin, and returns an EventResponse
func (p Plugin) Run(e Event) (EventResponse, error) {
	r := EventResponse{}

	eventToprocess := &e

	if len(e.Data) > maxMessageSize {
		copy := e.Copy()
		copy.Data = ""
		f, err := ioutil.TempFile(os.TempDir(), "pluggable")
		if err != nil {
			return r, errors.Wrap(err, "while creating temporary file")
		}
		if err := ioutil.WriteFile(f.Name(), []byte(e.Data), os.ModePerm); err != nil {
			return r, errors.Wrap(err, "while writing to temporary file")
		}
		copy.File = f.Name()
		eventToprocess = copy
		defer os.RemoveAll(f.Name())
	}

	k, err := eventToprocess.JSON()
	if err != nil {
		return r, errors.Wrap(err, "while marshalling event")
	}
	cmd := exec.Command(p.Executable, string(e.Name), k)
	cmd.Env = os.Environ()
	var b bytes.Buffer
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		r.Error = "error while executing plugin: " + err.Error() + string(b.String())
		return r, errors.Wrap(err, "while executing plugin: "+string(b.String()))
	}

	if err := json.Unmarshal(out, &r); err != nil {
		r.Error = err.Error()
		return r, errors.Wrap(err, "while unmarshalling response")
	}
	return r, nil
}
