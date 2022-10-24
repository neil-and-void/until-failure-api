package token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Credentials struct {
	ID    uint
	Name  string
	Email string
}

type Claims struct {
	Name string
	ID   uint
	jwt.StandardClaims
}

// signs a token
func Sign(c *Credentials, secret []byte, ttl time.Duration) string {
	claims := Claims{
		c.Name,
		c.ID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
			Issuer:    "neil:)",
			Subject:   c.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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

func Decode(tokenString string, secret []byte) (*Claims, error) {
	f := strings.Fields(tokenString)

	if len(f) != 2 || f[0] != "Bearer" {
		return nil, errors.New("Missing type \"Bearer\" in token string")
	}

	t, err := jwt.ParseWithClaims(f[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return &Claims{}, err
	}

	if claims, ok := t.Claims.(*Claims); ok && t.Valid {
		return claims, nil
	}

	return &Claims{}, nil
}
