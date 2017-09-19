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

	jwt "github.com/dgrijalva/jwt-go"
)

var signingSecret = []byte("should be in config file")

const badAuthMsg = "Authentication token not found."
const badSignMsg = "Unexpected authentication signing method: `%v`."

// Viewer represents the currently authenticated user.
type Viewer struct {
	ID    string `json:"uuid"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func makeAuthToken(user *User) (string, error) {
	claims := Viewer{
		user.ID,
		user.Email,
		jwt.StandardClaims{
			Issuer: "vaclav",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func decodeAuthToken(token string) (*Viewer, error) {
	result, err := jwtdecode(token)
	return result.Claims.(*Viewer), err
}

func isValidAuthToken(tokenString string) (bool, error) {
	token, err := jwtdecode(tokenString)
	if err != nil {
		return false, fmt.Errorf(badAuthMsg)
	}
	return token.Valid, nil
}

func jwtdecode(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf(badAuthMsg)
	}
	return jwt.ParseWithClaims(tokenString, &Viewer{}, checkAlgKey())
}

func checkAlgKey() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(badSignMsg, token.Header["alg"])
		}
		return signingSecret, nil
	}
}
