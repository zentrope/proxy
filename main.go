// Copyright 2017 Keith Irwin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

//-----------------------------------------------------------------------------

type AppMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
}

type App struct {
	Metadata AppMetadata `json:"metadata"`
	Icon     string      `json:"icon"`
	Context  string      `json:"context"`
}

// Remove the XML header from SVG and trim whitespace from either end.
func TrimSvg(svg string) string {
	reg, err := regexp.Compile("[<][?].*[?][>]")
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(reg.ReplaceAllString(svg, ""))
}

func findApps(dir string) ([]App, error) {

	var icons = make(map[string]string, 0)
	var metadata = make(map[string]AppMetadata, 0)

	setIcon := func(parent, p string) error {
		bytes, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		icons[parent] = TrimSvg(string(bytes))
		return nil
	}

	setMeta := func(parent, p string) error {
		bytes, err := ioutil.ReadFile(p)

		if err != nil {
			return err
		}

		var meta AppMetadata
		if err := json.Unmarshal(bytes, &meta); err != nil {
			return err
		}

		metadata[parent] = meta
		return nil
	}

	visit := func(p string, f os.FileInfo, err error) error {
		// Fragile because it's possible to find deeply nested svg and
		// metadata files using this general walker.

		if p == dir {
			return nil
		}

		parent := path.Base(path.Dir(p))

		if parent == path.Base(dir) {
			return nil
		}

		if path.Base(p) == "icon.svg" {
			if err := setIcon(parent, p); err != nil {
				return err
			}
		}

		if path.Base(p) == "metadata.js" {
			if err := setMeta(parent, p); err != nil {
				return err
			}
		}
		return nil
	}

	apps := make([]App, 0)

	err := filepath.Walk(dir, visit)
	if err != nil {
		return apps, err
	}

	for context, icon := range icons {
		apps = append(apps, App{
			Icon:     icon,
			Metadata: metadata[context],
			Context:  context,
		})
	}

	return apps, nil
}

//-----------------------------------------------------------------------------

func getPathContext(req *http.Request) string {
	return strings.Split(req.URL.Path, "/")[1]
}

func removePathContext(req *http.Request) string {
	context := getPathContext(req)
	path := req.URL.Path
	return strings.Replace(path, "/"+context, "", 1)
}

func makeContextDirector(routes RouteMap) func(req *http.Request) {
	return func(req *http.Request) {
		host := req.Host
		path := req.URL.Path

		req.URL.Scheme = "http"
		req.URL.Host = routes[getPathContext(req)]
		req.URL.Path = removePathContext(req)

		log.Printf("http://%v%v --> %v", host, path, req.URL.String())
	}
}

//-----------------------------------------------------------------------------

type RouteMap map[string]string

type ProxyConfig struct {
	StaticHandler http.Handler
	Routes        RouteMap
}

func (s ProxyConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := getPathContext(r)
	isApi := s.Routes[context] != ""

	// New reverse proxy handler each time so that we can dynamically
	// update the routing table in a thread-safe way.

	if isApi {
		p := &httputil.ReverseProxy{Director: makeContextDirector(s.Routes)}
		// Also takes a ModifyResponse function which you could use to
		// rewrite URL paths in the response if you assume the proxied app
		// assumes a root path.
		p.ServeHTTP(w, r)
		return
	}

	s.StaticHandler.ServeHTTP(w, r)
}

//-----------------------------------------------------------------------------

func main() {
	log.Println("Dynamic Proxy Experiment")

	apps, err := findApps("./public")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", apps)

	routes := RouteMap{
		"api":  "localhost:10001",
		"api2": "localhost:10002",
	}

	proxy := ProxyConfig{
		StaticHandler: http.FileServer(http.Dir("./public")),
		Routes:        routes,
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	log.Fatal(server.ListenAndServe())
}
