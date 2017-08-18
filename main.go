// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

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

//----

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

func main() {
	log.Println("Dynamic Proxy Experiment")

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
