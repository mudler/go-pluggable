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
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

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
		var m *Manager

		BeforeEach(func() {
			pluginFile, err = ioutil.TempFile(os.TempDir(), "tests")
			Expect(err).Should(BeNil())
			defer os.Remove(pluginFile.Name()) // clean up
			pluginFile2, err = ioutil.TempFile(os.TempDir(), "tests")
			Expect(err).Should(BeNil())
			defer os.Remove(pluginFile2.Name()) // clean up
			m = NewManager([]EventType{})
		})

		It("autoload plugins", func() {
			temp, err := ioutil.TempDir(os.TempDir(), "autoload")
			Expect(err).Should(BeNil())

			d1 := []byte("#!/bin/bash\necho \"{ \\\"state\\\": \\\"$1\\\" }\"\n")
			err = ioutil.WriteFile(filepath.Join(temp, "test-foo"), d1, 0550)
			Expect(err).Should(BeNil())

			m.Autoload("test", temp)
			m.Events = []EventType{PackageInstalled}
			m.Register()

			var received map[string]string
			var resp *EventResponse

			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})
			m.Publish(PackageInstalled, map[string]string{"foo": "bar"})

			Expect(resp).ToNot(BeNil())
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(resp.State).Should(Equal(string(PackageInstalled)))
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
			m.Register()

			var received map[string]string
			var resp *EventResponse

			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})
			m.Publish(PackageInstalled, map[string]string{"foo": "bar"})

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
			m.Register()

			var received map[string]string
			var resp *EventResponse

			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})
			m.Publish(PackageInstalled, map[string]string{"foo": "bar"})

			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(resp.State).Should(Equal(string(PackageInstalled)))
		})

		It("gets the json event payload", func() {
			d1 := []byte("#!/bin/bash\necho $2\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}

			foo := map[string]string{"foo": "bar"}
			m.Register()

			var received map[string]string
			var resp *EventResponse

			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
				r.Unmarshal(&received)
			})
			m.Publish(PackageInstalled, foo)
			Expect(resp.Errored()).ToNot(BeTrue())
			Expect(received).Should(Equal(foo))
		})

		It("gets the plugin", func() {
			d1 := []byte("#!/bin/bash\necho $2\n")
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Register()

			foo := map[string]string{"foo": "bar"}

			var received map[string]string
			var receivedPlugin *Plugin
			var resp *EventResponse
			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
				receivedPlugin = p
				r.Unmarshal(&received)
			})
			m.Publish(PackageInstalled, foo)
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
			m.Register()

			foo := map[string]string{"foo": "bar"}

			var received []map[string]string
			var receivedPlugin []*Plugin
			var resp []EventResponse
			mu := sync.Mutex{}

			f := func(p *Plugin, r *EventResponse) {
				mu.Lock()
				defer mu.Unlock()
				resp = append(resp, *r)
				receivedPlugin = append(receivedPlugin, p)
				var rec map[string]string
				r.Unmarshal(&rec)
				received = append(received, rec)
			}
			m.Response(PackageInstalled, f)
			m.Publish(PackageInstalled, foo)

			Eventually(func() int {
				mu.Lock()
				defer mu.Unlock()
				return len(resp)
			}, 100*time.Second).Should(Equal(2))

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
			m.Register()

			foo := map[string]string{"foo": "bar"}
			var received map[string]string
			var resp *EventResponse
			mu := sync.Mutex{}
			f := func(p *Plugin, r *EventResponse) {
				mu.Lock()
				resp = r
				r.Unmarshal(&received)
				mu.Unlock()
			}
			m.Response(PackageInstalled, f)
			go m.Publish(PackageInstalled, foo)
			go m.Publish(PackageInstalled, foo)
			go m.Publish(PackageInstalled, foo)
			go m.Publish(PackageInstalled, foo)

			Eventually(func() map[string]string {

				mu.Lock()
				defer mu.Unlock()
				return received
			}).Should(Equal(foo))

			mu.Lock()

			Expect(resp.Errored()).ToNot(BeTrue())

			mu.Unlock()
		})

		It("Writes the data to a file when it's too big", func() {
			d1 := []byte(`#!/bin/bash
echo "{ \"data\": \"$(echo $2 | base64 -w0)\" }"`)
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Register()

			foo := map[string]string{"foo": randStringRunes((1 << 13) + 1)}

			var resp *EventResponse
			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
			})
			m.Publish(PackageInstalled, foo)
			b, err := base64.StdEncoding.DecodeString(resp.Data)

			Expect(string(b)).To(ContainSubstring("{\"name\":\"package.install\",\"data\":\"\",\"file\":\"/tmp/pluggable"))
			Expect(err).Should(BeNil())

		})

		It("The file that is written has the event content", func() {
			d1 := []byte(`#!/bin/bash
			data="$(cat $(echo $2 | jq -r .file))"
			jq --arg key0   'data' \
			   --arg value0 "$data" \
				'. | .[$key0]=$value0' \
				<<<'{}'`)
			err := ioutil.WriteFile(pluginFile.Name(), d1, 0550)
			Expect(err).Should(BeNil())

			m.Plugins = []Plugin{{Name: "test", Executable: pluginFile.Name()}}
			m.Events = []EventType{PackageInstalled}
			m.Register()

			foo := map[string]string{"foo": randStringRunes((1 << 13) + 1)}

			var resp *EventResponse
			m.Response(PackageInstalled, func(p *Plugin, r *EventResponse) {
				resp = r
			})
			m.Publish(PackageInstalled, foo)
			res := map[string]string{}
			err = resp.Unmarshal(&res)
			Expect(err).Should(BeNil())

			Expect(res["foo"]).ToNot(BeEmpty())
		})
	})
})

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
