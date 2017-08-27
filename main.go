//-----------------------------------------------------------------------------
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
//-----------------------------------------------------------------------------

package main

import (
	"log"
	"net/http"

	"github.com/zentrope/proxy/internal"
)

func main() {

	log.Println("Dynamic Proxy Experiment")

	appDir := "./public"
	hostDir := "./client"

	proxy := internal.NewProxyServer(appDir, hostDir)
	proxy.AddRoute("api", "localhost:10001")
	proxy.TestConnections()

	server := http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	log.Fatal(server.ListenAndServe())
}
