// Copyright © 2020 Ettore Di Giacinto <mudler@mocaccino.org>
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
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/mudler/go-pluggable"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	PackageInstalled EventType = "package.install"
)

var _ = Describe("Plugin", func() {
	Context("event subscription", func() {
		var pluginFile *os.File

		var pluginFile2 *os.File

		var err error
		var b *Bus
		var m *Manager

		BeforeEach(func() {
			pluginFile, err = ioutil.TempFile(os.TempDir(), "tests")
			Expect(err).Should(BeNil())
			defer os.Remove(pluginFile.Name()) // clean up
			pluginFile2, err = ioutil.TempFile(os.TempDir(), "tests")
			Expect(err).Should(BeNil())
			defer os.Remove(pluginFile2.Name()) // clean up
			b = NewBus()
			m = &Manager{}
		})

		It("autoload plugins", func() {
			temp, err := ioutil.TempDir(os.TempDir(), "autoload")
			Expect(err).Should(BeNil())

			d1 := []byte("#!/bin/bash\necho \"{ \\\"state\\\": \\\"$1\\\" }\"\n")
			err = ioutil.WriteFile(filepath.Join(temp, "test-foo"), d1, 0550)
			Expect(err).Should(BeNil())

			m.Autoload("test", temp)
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			ev, err := NewEvent(PackageInstalled, map[string]string{"foo": "bar"})
			Expect(err).Should(BeNil())

			var received map[string]string
			var resp *EventResponse

			b.Publish(ev,
				func(r *EventResponse) {
					resp = r
					r.Unmarshal(&received)
				})

			Expect(resp).ToNot(BeNil())
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(resp.State).Should(Equal(string(PackageInstalled)))
		})

		When("using manager publish", func() {
			It("propagates events", func() {
				temp, err := ioutil.TempDir(os.TempDir(), "autoload")
				Expect(err).Should(BeNil())

				d1 := []byte("#!/bin/bash\necho \"{ \\\"state\\\": \\\"$1\\\" }\"\n")
				err = ioutil.WriteFile(filepath.Join(temp, "test-foo"), d1, 0550)
				Expect(err).Should(BeNil())

				m.Autoload("test", temp)
				m.Events = []EventType{PackageInstalled}
				m.Bus = b
				m.Register()

				var received map[string]string
				var resp *EventResponse
				m.Publish(PackageInstalled, map[string]string{"foo": "bar"}, func(r *EventResponse) {
					resp = r
					r.Unmarshal(&received)
				})
				Expect(resp).ToNot(BeNil())
				Expect(resp.Errored()).ToNot(BeTrue())
				Expect(resp.State).Should(Equal(string(PackageInstalled)))
			})
		})

		It("loads plugins", func() {
			temp, err := ioutil.TempDir(os.TempDir(), "autoload")
			Expect(err).Should(BeNil())

			d1 := []byte("#!/bin/bash\necho \"{ \\\"state\\\": \\\"$1\\\" }\"\n")
			err = ioutil.WriteFile(filepath.Join(temp, "test-foo"), d1, 0550)
			Expect(err).Should(BeNil())
			os.Setenv("PATH", os.Getenv("PATH")+":"+temp)
			m.Load("test-foo")
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			ev, err := NewEvent(PackageInstalled, map[string]string{"foo": "bar"})
			Expect(err).Should(BeNil())

			var received map[string]string
			var resp *EventResponse

			b.Publish(ev, func(r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})

			Expect(resp).ToNot(BeNil())
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(resp.State).Should(Equal(string(PackageInstalled)))
		})

		It("gets the json event name", func() {
			d1 := []byte("#!/bin/bash\necho \"{ \\\"state\\\": \\\"$1\\\" }\"\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			ev, err := NewEvent(PackageInstalled, map[string]string{"foo": "bar"})
			Expect(err).Should(BeNil())

			var received map[string]string
			var resp *EventResponse

			b.Publish(ev, func(r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(resp.State).Should(Equal(string(PackageInstalled)))
		})

		It("gets the json event payload", func() {
			d1 := []byte("#!/bin/bash\necho $2\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			foo := map[string]string{"foo": "bar"}
			ev, err := NewEvent(PackageInstalled, foo)
			Expect(err).Should(BeNil())

			var received map[string]string
			var resp *EventResponse

			b.Publish(ev, func(r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(received).Should(Equal(foo))
		})

		It("gets the plugin", func() {
			d1 := []byte("#!/bin/bash\necho $2\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			foo := map[string]string{"foo": "bar"}
			ev, err := NewEvent(PackageInstalled, foo)
			Expect(err).Should(BeNil())

			var received map[string]string
			var receivedPlugin *Plugin
			var resp *EventResponse

			b.Publish(ev, func(p *Plugin, r *EventResponse) {
				resp = r
				receivedPlugin = p
				r.Unmarshal(&received)
			})
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(received).Should(Equal(foo))
			Expect(receivedPlugin.Name).Should(Equal("test"))
		})

		It("gets multiple plugin responses", func() {
			d1 := []byte("#!/bin/bash\necho $2\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())
			err = ioutil.WriteFile(pluginFile2.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()},
				{Name: "test2", Executable: pluginFile2.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			foo := map[string]string{"foo": "bar"}
			ev, err := NewEvent(PackageInstalled, foo)
			Expect(err).Should(BeNil())

			var received []map[string]string
			var receivedPlugin []*Plugin
			var resp []*EventResponse

			b.Publish(ev, func(p *Plugin, r *EventResponse) {
				resp = append(resp, r)
				receivedPlugin = append(receivedPlugin, p)
				var rec map[string]string
				r.Unmarshal(&rec)
				received = append(received, rec)
			})

			Expect(len(resp)).To(Equal(2))

			for _, r := range resp {
				Expect(r.Errored()).ToNot(BeTrue())
			}
			for _, r := range received {
				Expect(r).Should(Equal(foo))
			}
			Expect(receivedPlugin).To(ContainElement(&Plugin{Name: "test2", Executable: pluginFile2.Name()}))
			Expect(receivedPlugin).To(ContainElement(&Plugin{Name: "test", Executable: pluginFile.Name()}))
		})
		It("is concurrent safe", func() {
			d1 := []byte("#!/bin/bash\necho $2\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Subscribe(b)

			foo := map[string]string{"foo": "bar"}
			ev, err := NewEvent(PackageInstalled, foo)
			Expect(err).Should(BeNil())

			var received map[string]string
			var resp *EventResponse
			b.Listen(ev.ResponseEventName("results"), func(r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})

			go b.Publish(ev)
			go b.Publish(ev)
			go b.Publish(ev)
			go b.Publish(ev)

			Eventually(func() map[string]string {
				return received
			}).Should(Equal(foo))

			Expect(resp.Errored()).ToNot(BeTrue())
		})

	})
})
