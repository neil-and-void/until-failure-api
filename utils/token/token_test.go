package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Parallel()

	c := Credentials{
		ID:    12,
		Email: "test@test.com",
		Name:  "testname",
	}
	secret := "somesecret"
	var ttl time.Duration = 168 // days

	t.Run("Successfully sign and decode a token", func(t *testing.T) {
		tkn := Sign(&c, []byte(secret), ttl)

		claims, err := Decode("Bearer "+tkn, []byte(secret))

		assert.Nil(t, err, "Error decoding token")
		assert.Equal(t, claims.Subject, "test@test.com")
		assert.Equal(t, claims.Name, "testname")
	})

	t.Run("Fail to decode a tampered token", func(t *testing.T) {
		tkn := Sign(&c, []byte(secret), ttl)
		tamperedToken := tkn + "hehehe"

		_, err := Decode(tamperedToken, []byte("Bearer "+secret))
		assert.NotNil(t, err, "There should be an error decoding")
	})

	t.Run("Fail to validate an expired token", func(t *testing.T) {
		tkn := Sign(&c, []byte(secret), -5) // 5 hours in the past from now

		_, err := Decode(tkn, []byte("Bearer "+secret))

		assert.NotNil(t, err, "Should be an error decoding a token")
	})
}
