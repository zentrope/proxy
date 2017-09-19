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

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// A User represents a user of the system.
type User struct {
	ID       string
	Email    string
	Password string
}

type appStoreSku struct {
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

// Database represents the abstraction for storing application data.
type Database struct {
	users []*User
	skus  []*appStoreSku
}

// NewDatabase returns a database abstraction for storing application data.
func NewDatabase() *Database {
	return &Database{
		users: []*User{newUser("test@example.com", "test1234")},
	}
}

// Start the database service.
func (db *Database) Start() {
	log.Println("Starting database.")
}

// Stop the database service
func (db *Database) Stop() {
	log.Println("Stopping database.")
}

func newUser(email, password string) *User {
	passcode, err := encryptPassword(password)
	if err != nil {
		log.Fatalf("Unable to encrypt password: %v", err)
	}
	return &User{
		ID:       mkUUID(),
		Email:    email,
		Password: passcode,
	}
}

func (db *Database) findUser(email, password string) (*User, error) {
	test := strings.ToLower(email)
	for _, u := range db.users {
		if strings.ToLower(u.Email) == test && validPassword(password, u.Password) {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (db *Database) findSKU(xrn string) (*appStoreSku, error) {
	for _, sku := range db.skus {
		if sku.XRN == xrn {
			return sku, nil
		}
	}
	return nil, fmt.Errorf("app '%s' not found", xrn)
}

func (db *Database) appSkus() []*appStoreSku {
	// Return skus, but remove some of the data

	skus := make([]*appStoreSku, 0)
	for _, s := range db.skus {
		var newSku appStoreSku
		newSku = *s
		newSku.Context = ""
		newSku.DownloadURL = ""
		skus = append(skus, &newSku)
	}

	return skus
}

func (db *Database) setSKUs(skus []*appStoreSku) {
	db.skus = skus
}

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

func mkUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}
