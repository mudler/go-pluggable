// Copyright © 2021 Ettore Di Giacinto <mudler@mocaccino.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

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
			err = json.Unmarshal([]byte(b.String()), resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Data).To(Equal("true"))
		})
	})
})
