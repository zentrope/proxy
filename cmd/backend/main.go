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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/zentrope/proxy/cmd/backend/data"
)

type ServerConfig struct {
	port    string
	message string
}

func writeFile(w http.ResponseWriter, content string) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Write([]byte(content))
}

var routeMap = map[string]string{
	"scan":     data.Scans,
	"schedule": data.Schedule,
}

func (s ServerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := r.Header.Get("X-Proxy-Context")
	log.Printf("request: (%v) -> %v", prefix, r.URL)

	context := strings.Split(r.URL.Path, "/")[1]
	route := routeMap[context]

	if context == "kill9" {
		os.Exit(9)
	}

	if route != "" {
		writeFile(w, route)
		return
	}

	doc := `<!doctype html>
	 <html>
		 <head>
			 <title>%v</title>
		 </head>
		 <body>
			<h1>%v</h1>
			<ul>
				<li><a href="/%s/scan">Isolinear matrix scans</a></li>
				<li><a href="/%s/schedule">Engineering deck maintenance schedules</a></li>
			</ul>
		 </body>
		</html>`
	id := fmt.Sprintf("%v (port: %v)", s.message, s.port)
	fmt.Fprintf(w, doc, id, id, prefix, prefix)
}

func main() {

	// The goal is to create a simple backend server we can use to test
	// a dynamic proxy.

	port := flag.String("port", "10001", "Port")
	msg := flag.String("message", "Starship Maintenance", "Message to display.")

	flag.Parse()

	log.Printf("Test Server\n")
	log.Printf(" msg:  %v\n", *msg)
	log.Printf(" port: %v\n", *port)

	config := ServerConfig{*port, *msg}

	server := http.Server{
		Addr:    "127.0.0.1:" + config.port,
		Handler: config,
	}

	log.Fatal(server.ListenAndServe())
}
