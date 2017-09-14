//
// Copyright (C) 2017 Keith Irwin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
