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

package internal

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

//-----------------------------------------------------------------------------

type Client struct {
	token string
	conn  *websocket.Conn
}

func NewClient(token string, conn *websocket.Conn) *Client {
	return &Client{
		token: token,
		conn:  conn,
	}
}

func (client *Client) Send(msg interface{}) error {
	return websocket.WriteJSON(client.conn, msg)
}

func (client *Client) SendAck(command string) error {
	type ackMsg struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}
	msg := ackMsg{"ack", command}
	return client.Send(msg)
}

//-----------------------------------------------------------------------------

type ClientHub struct {
	clients []*Client
	mutex   sync.Mutex
}

func NewClientHub() *ClientHub {
	return &ClientHub{
		mutex: sync.Mutex{},
	}
}

func (hub *ClientHub) Start() {
	log.Println("Starting client hub.")
}

func (hub *ClientHub) Stop() {
	log.Println("Stopping client hub.")
}

func (hub *ClientHub) SendAck(token, command string) error {
	for _, c := range hub.clients {
		if c.token == token {
			return c.SendAck(command)
		}
	}
	return fmt.Errorf("Unable to find client to ack.")
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
		if err := client.Send(commandRefresh); err != nil {
			log.Printf("ERROR: Unable to write to socket.")
		}
	}
}
