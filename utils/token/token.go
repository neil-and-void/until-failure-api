package token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Credentials struct {
	ID    uint
	Name  string
	Email string
}

// takes claims and puts it into a Credentials struct representing the user
func ClaimsToStruct(c jwt.MapClaims) *Credentials {
	// need to convert interface{} to uint
	id, ok := c["ID"].(uint)
	if !ok {
		return &Credentials{}
	}
	name := fmt.Sprintf("%v", c["name"])
	email := fmt.Sprintf("%v", c["email"])

	return &Credentials{
		ID: id,
		Name: name,
		Email: email,
	}
}

// signs a token
func Sign(c Credentials, secret []byte, dtl int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   c.Name,
		"email": c.Email,
		"exp":   time.Now().UTC().AddDate(0, 0, dtl).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(secret)

	if err != nil {
		panic(err)
	}

	return tokenString
}

func Validate(tokenString string, secret []byte) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return secret, nil
	})
	if err != nil {
		panic(err)
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true
	} else {
		return false
	}
}

func Decode(tokenString string, secret []byte) (jwt.MapClaims, error) {
	f := strings.Fields(tokenString)

	if len(f) != 2 || f[0] != "Bearer" {
		return nil,  errors.New("Missing type \"Bearer\" in token string")
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(f[1], claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}
	return claims, nil
}
