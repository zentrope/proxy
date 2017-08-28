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
	"fmt"
	"log"

	jwt "github.com/dgrijalva/jwt-go"
)

var SECRET = []byte("should be in config file")

const BAD_AUTH_MSG = "Not found."

type ViewerClaims struct {
	Id    string `json:"uuid"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func MakeAuthToken(user *User) (string, error) {

	claims := ViewerClaims{
		user.Id,
		user.Email,
		jwt.StandardClaims{
			Issuer: "vaclav",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SECRET)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func IsValidAuthToken(tokenString string) (bool, error) {

	if tokenString == "" {
		log.Printf(" [x] auth.error: Token not found.")
		return false, fmt.Errorf(BAD_AUTH_MSG)
	}

	token, err := jwt.ParseWithClaims(tokenString, &ViewerClaims{}, checkAlgKey())

	if err != nil {
		log.Printf(" [x] auth.error: %v", err)
		return false, fmt.Errorf(BAD_AUTH_MSG)
	}

	return token.Valid, nil
}

func checkAlgKey() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(" [x] auth.error: unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET, nil
	}
}
