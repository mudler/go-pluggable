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

package pluggable_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	. "github.com/mudler/go-pluggable"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginFactory", func() {
	Context("creating plugins", func() {
		factory := NewPluginFactory()

		BeforeEach(func() {
			factory = NewPluginFactory()
		})

		It("reacts to events", func() {
			b := bytes.NewBufferString("")
			payload := &Event{Name: "foo", Data: "bar"}

			payloadDat, err := json.Marshal(payload)
			Expect(err).ToNot(HaveOccurred())
			factory.Add("foo", func(e *Event) EventResponse { return EventResponse{State: "foo", Data: fmt.Sprint(e.Data == "bar")} })
			err = factory.Run("foo", string(payloadDat), b)
			Expect(err).ToNot(HaveOccurred())

			Expect(b.String()).ToNot(BeEmpty())
			resp := &EventResponse{}
			err = json.Unmarshal(b.Bytes(), resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Data).To(Equal("true"))
		})
	})
})
