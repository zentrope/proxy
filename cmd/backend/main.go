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
)

//-----------------------------------------------------------------------------

const Scans = `
[
	{
		"process": "proc:gl:904e331d-1853-4472-8694-79410cbea625",
		"isolinear_matrix": "10.0.1/24",
		"start": "2017-05-05 11:05:00",
		"stop": "2017-07-04 04:05:00",
		"schedule": {
			"month": "*",
			"dayOfWeek": "*",
			"hours": "*",
			"minutes": "*",
			"seconds": "*",
			"dayOfMonth": "*"
		}
	},
	{
		"process": "proc:gl:cc109dae-eb7e-4e99-b848-70c7cdb56829",
		"isolinear_matrix": "10.0.2/24",
		"start": "2015-09-12 11:05:00",
		"stop": "2018-11-04 04:05:00",
		"schedule": {
			"month": "5,6,7",
			"dayOfWeek": "4,5",
			"hours": "4",
			"minutes": "20",
			"seconds": "*",
			"dayOfMonth": "*"
		}
	},
	{
		"process": "proc:gl:bf8f8928-9abb-46b7-b5ed-b5d5dad0bc5e",
		"isolinear_matrix": "10.0.3/24",
		"start": "2017-05-05 11:05:00",
		"stop": "2017-07-04 04:05:00",
		"schedule": {
			"month": "*",
			"dayOfWeek": "4",
			"hours": "7",
			"minutes": "17",
			"seconds": "*",
			"dayOfMonth": "*"
		}
	}
]
`

const Schedule = `
[
	{
		"id": "1",
		"status": "active",
		"job_id": "0001",
		"name": "Flush Containment System",
		"description": "Purge ionized dilithium particles Sundays at 3AM.",
		"process": "proc:gl:904e331d-1853-4472-8694-79410cbea625",
		"schedule": {
			"minute": "15",
			"hour": "3",
			"month": "1",
			"year": "*",
			"date": "*",
			"day": "7"
		}
	},
	{
		"id": "2",
		"status": "active",
		"job_id": "0002",
		"name": "Phase Inverter Report",
		"description": "Run the phase inverter report at 5:10 AM every morning.",
		"component": "proc:gl:904e331d-1853-4472-8694-79410cbea625",
		"schedule": {
			"minute": "10",
			"hour": "5",
			"month": "*",
			"year": "*",
			"date": "3",
			"day": "*"
		}
	},
	{
		"id": "3",
		"status": "active",
		"job_id": "0003",
		"name": "Clean Holo Emitter Arrays",
		"description": "A clean holo emitter array is a lovely thing.",
		"component": "proc:gl:bf8f8928-9abb-46b7-b5ed-b5d5dad0bc5e",
		"schedule": {
			"minute": "10",
			"hour": "5",
			"month": "*",
			"year": "*",
			"date": "*",
			"day": "21"
		}
	}
]
`

//-----------------------------------------------------------------------------

type ServerConfig struct {
	port    string
	message string
}

func writeFile(w http.ResponseWriter, content string) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Write([]byte(content))
}

var routeMap = map[string]string{
	"scan":     Scans,
	"schedule": Schedule,
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
