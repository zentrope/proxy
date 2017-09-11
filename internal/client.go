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

package internal

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	token string
	conn  *websocket.Conn
}

type ClientHub struct {
	clients []*Client
	mutex   sync.Mutex
}

func NewClientHub() *ClientHub {
	return &ClientHub{
		mutex: sync.Mutex{},
	}
}

func NewClient(token string, conn *websocket.Conn) *Client {
	return &Client{
		token: token,
		conn:  conn,
	}
}

func (hub *ClientHub) Start() {
	log.Println("Starting client hub.")
}

func (hub *ClientHub) Stop() {
	log.Println("Stopping client hub.")
}

func (hub *ClientHub) Add(client *Client) *Client {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.clients = append(hub.clients, client)
	log.Printf("- %v attached client(s)", len(hub.clients))
	return client
}

func (hub *ClientHub) Delete(client *Client) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	clients := make([]*Client, 0)
	for _, c := range hub.clients {
		if c.conn != client.conn {
			clients = append(clients, c)
		}
	}
	hub.clients = clients
	log.Printf("- %v attached client(s)", len(hub.clients))
}

type simpleNotification struct {
	Type string `json:"type"`
}

var commandRefresh = simpleNotification{"refresh"}

func (hub *ClientHub) NotifyRefresh() {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	for _, client := range hub.clients {
		if err := websocket.WriteJSON(client.conn, commandRefresh); err != nil {
			log.Printf("ERROR: Unable to write to socket.")
		}
	}
}
