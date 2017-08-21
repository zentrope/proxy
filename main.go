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
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/zentrope/proxy/server"
)

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

func (s ProxyConfig) IsApi(r *http.Request) bool {
	return s.Routes[getPathContext(r)] != ""
}

//-----------------------------------------------------------------------------

type RouteMap map[string]string

type ProxyConfig struct {
	Applications  *server.Applications
	Routes        RouteMap
	StaticHandler http.Handler
}

func logRequest(r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL.Path)
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func (s ProxyConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	// Localhost development.
	setCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method == "HEAD" {
		return
	}

	if s.IsApi(r) {
		//
		// New reverse proxy handler each time so that we can dynamically
		// update the routing table in a thread-safe way.
		//
		p := &httputil.ReverseProxy{Director: makeContextDirector(s.Routes)}
		//
		// Also takes a ModifyResponse function which you could use to
		// rewrite URL paths in the response if you assume the proxied app
		// assumes a root path.
		//
		p.ServeHTTP(w, r)
		return
	}

	switch getPathContext(r) {

	case "shell":
		s.Applications.Reload()
		json, err := s.Applications.AsJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, json)
		return

	default:
		s.StaticHandler.ServeHTTP(w, r)
	}
}

//-----------------------------------------------------------------------------

func main() {
	log.Println("Dynamic Proxy Experiment")

	appDir := "./public"

	routes := RouteMap{
		"api":  "localhost:10001",
		"api2": "localhost:10002",
	}

	proxy := ProxyConfig{
		StaticHandler: http.FileServer(http.Dir(appDir)),
		Applications:  server.NewApplications(appDir),
		Routes:        routes,
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	log.Fatal(server.ListenAndServe())
}
