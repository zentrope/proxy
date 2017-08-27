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

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       string
	Email    string
	Password string
}

type Database struct {
	Users []*User
}

func NewDatabase() *Database {
	return &Database{
		Users: []*User{NewUser("test@example.com", "test1234")},
	}
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
