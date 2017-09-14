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

//-----------------------------------------------------------------------------
// A fake database for exploring auth related issues.
//-----------------------------------------------------------------------------

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

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
}

func NewDatabase() *Database {
	return &Database{
		Users: []*User{NewUser("test@example.com", "test1234")},
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
