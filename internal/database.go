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

//-----------------------------------------------------------------------------
// A fake database for exploring auth related issues.
//-----------------------------------------------------------------------------

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       string
	Email    string
	Password string
}

type AppStoreSku struct {
	XRN         string `json:"xrn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Context     string `json:"context,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	IsInstalled bool   `json:"is_installed"`
}

type ClientConn struct {
	token string
	conn  *websocket.Conn
}

type Database struct {
	Users   []*User
	Skus    []*AppStoreSku
	Clients []*ClientConn
	mutex   sync.Mutex
}

func NewDatabase() *Database {
	return &Database{
		Users: []*User{NewUser("test@example.com", "test1234")},
		mutex: sync.Mutex{},
	}
}

func (db *Database) Start() {
	log.Println("Starting database.")
}

func (db *Database) Stop() {
	log.Println("Stopping database.")
}

func NewUser(email, password string) *User {
	passcode, err := encryptPassword(password)
	if err != nil {
		log.Fatal("Unable to encrypt password: %v", err)
	}
	return &User{
		Id:       mkUuid(),
		Email:    email,
		Password: passcode,
	}
}

func (db *Database) FindUser(email, password string) (*User, error) {
	test := strings.ToLower(email)
	for _, u := range db.Users {
		if strings.ToLower(u.Email) == test && validPassword(password, u.Password) {
			return u, nil
		}
	}
	return nil, errors.New("User not found.")
}

func (db *Database) FindSKU(xrn string) (*AppStoreSku, error) {
	for _, sku := range db.Skus {
		if sku.XRN == xrn {
			return sku, nil
		}
	}
	return nil, fmt.Errorf("App '%s' not found.", xrn)
}

func (db *Database) SKUs() []*AppStoreSku {
	// Return skus, but remove some of the data

	skus := make([]*AppStoreSku, 0)
	for _, s := range db.Skus {
		var newSku AppStoreSku
		newSku = *s
		newSku.Context = ""
		newSku.DownloadURL = ""
		skus = append(skus, &newSku)
	}

	return skus
}

func (db *Database) SetSKUs(skus []*AppStoreSku) {
	db.Skus = skus
}

func (db *Database) AddClient(token string, conn *websocket.Conn) *ClientConn {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	client := &ClientConn{token, conn}
	db.Clients = append(db.Clients, client)
	return client
}

func (db *Database) DeleteClient(conn *websocket.Conn) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	clients := make([]*ClientConn, 0)
	for _, client := range db.Clients {
		if client.conn != conn {
			clients = append(clients, client)
		}
	}
	db.Clients = clients
}

type SimpleNotification struct {
	Type string `json:"type"`
}

func (db *Database) NotifyRefresh() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	refresh := SimpleNotification{"refresh"}
	for _, client := range db.Clients {
		if err := websocket.WriteJSON(client.conn, refresh); err != nil {
			log.Printf("ERROR: Unable to write to socket.")
		}
	}
}

//-----------------------------------------------------------------------------

func validPassword(password, hash string) bool {
	decoded, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(decoded, []byte(password))
	if err != nil {
		return false
	}
	return true
}

func encryptPassword(password string) (string, error) {
	raw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", raw), nil
}

func mkUuid() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}
