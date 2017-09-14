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

	clients := internal.NewClientHub()
	database := internal.NewDatabase()
	appstore := internal.NewAppStore(appStoreUrl, database)
	commander := internal.NewCommandProcessor(appDir, database, clients)

	proxy := internal.NewProxyServer(appDir, hostDir, database, commander, clients)
	proxy.AddRoute("api", "127.0.0.1:10001")

	clients.Start()
	database.Start()
	commander.Start()
	appstore.Start()
	proxy.Start()

	blockUntilShutdownThenDo(func() {
		log.Println("Shutdown")
		proxy.Stop()
		appstore.Stop()
		commander.Stop()
		database.Stop()
		clients.Stop()
	})

	log.Println("System halt.")
}
