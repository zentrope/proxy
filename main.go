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
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/zentrope/proxy/internal"
)

//-----------------------------------------------------------------------------

func getPathContext(req *http.Request) string {
	context := strings.Split(req.URL.Path, "/")[1]
	// If the context contains a ".", assume it's a data file at the top
	// of the directory.
	if strings.Index(context, ".") != -1 {
		return ""
	}
	return context
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
		context := getPathContext(req)

		req.URL.Scheme = "http"
		req.URL.Host = routes[context]
		req.URL.Path = removePathContext(req)

		// So that back-ends can prefix URLs to get back here.
		//req.Header.Set("X-Proxy-Context", "http://"+req.Host+"/"+context)
		req.Header.Set("X-Proxy-Context", context)

		log.Printf("http://%v%v --> %v", host, path, req.URL.String())
	}
}

func (s ProxyConfig) IsApi(r *http.Request) bool {
	return s.Routes[getPathContext(r)] != ""
}

//-----------------------------------------------------------------------------

type RouteMap map[string]string

type ProxyConfig struct {
	Applications   *internal.Applications
	Routes         RouteMap
	RootAppHandler http.Handler
	StaticHandler  http.Handler
}

func logRequest(r *http.Request) {
	log.Printf("[ %-12v ] -- %v %v", getPathContext(r), r.Method, r.URL.Path)
}

func isViableConnection(addr string) bool {

	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err != nil {
		log.Printf("Connection error: %v", err)
		return false
	}

	defer conn.Close()

	return true
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
		p := &httputil.ReverseProxy{
			Director: makeContextDirector(s.Routes),
			ModifyResponse: func(res *http.Response) error {
				res.Header.Set("X-Proxy-Context", getPathContext(r))
				return nil
			},
		}

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

	case "":
		s.RootAppHandler.ServeHTTP(w, r)
		return

	default:
		s.StaticHandler.ServeHTTP(w, r)
	}
}

//-----------------------------------------------------------------------------

func main() {
	log.Println("Dynamic Proxy Experiment")

	appDir := "./public"
	hostDir := "./client"

	routes := RouteMap{
		"api": "localhost:10001",
		// "api2": "localhost:10002",
	}

	for k, v := range routes {
		if !isViableConnection(v) {
			log.Printf("WARNING: ROUTE '/%v' CANNOT CONNECT TO '%v'", k, v)
		}
	}

	proxy := ProxyConfig{
		StaticHandler:  http.FileServer(http.Dir(appDir)),
		RootAppHandler: http.FileServer(http.Dir(hostDir)),
		Applications:   internal.NewApplications(appDir),
		Routes:         routes,
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	log.Fatal(server.ListenAndServe())
}
