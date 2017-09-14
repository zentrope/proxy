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
