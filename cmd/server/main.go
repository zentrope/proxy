// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
