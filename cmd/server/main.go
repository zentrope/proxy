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
)

type ServerConfig struct {
	port    string
	message string
}

func (s ServerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doc := `<!doctype html>
	 <html>
		 <head>
			 <title>%v</title>
		 </head>
		 <body>
			<h1>%v</h1>
		 </body>
		</html>`
	log.Printf("request: %v", r.URL)
	id := fmt.Sprintf("%v (port: %v)", s.message, s.port)
	fmt.Fprintf(w, doc, id, id)
}

func main() {

	// The goal is to create a simple backend server we can use to test
	// a dynamic proxy.

	port := flag.String("port", "10001", "Port")
	msg := flag.String("message", "Hello World", "Message to display.")

	flag.Parse()

	fmt.Printf("Test Server\n")
	fmt.Printf(" port: %v\n", *port)
	fmt.Printf(" msg:  %v\n", *msg)

	config := ServerConfig{*port, *msg}

	server := http.Server{
		Addr:    "127.0.0.1:" + config.port,
		Handler: config,
	}

	log.Fatal(server.ListenAndServe())
}
