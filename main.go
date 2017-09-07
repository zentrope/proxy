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
	"os"
	"os/signal"
	"syscall"

	"github.com/zentrope/proxy/internal"
)

func blockUntilShutdownThenDo(fn func()) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP)
	v := <-sigChan
	log.Printf("Signal: %v\n", v)
	fn()
}

func main() {

	log.Println("Dynamic Proxy Experiment")

	appDir := "./public"
	hostDir := "./client"
	appStoreUrl := "http://localhost:60001"

	appstore := internal.NewAppStore(appStoreUrl)
	commander := internal.NewCommandProcessor(appstore)

	proxy := internal.NewProxyServer(appDir, hostDir, appstore, commander)
	proxy.AddRoute("api", "127.0.0.1:10001")

	go proxy.Start()
	go appstore.Start()

	blockUntilShutdownThenDo(func() {
		log.Println("Shutdown")
		appstore.Stop()
		proxy.Stop()
	})

	log.Println("System halt.")
}
