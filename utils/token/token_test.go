package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Parallel()

	c := Credentials{
		ID: 12,
		Email: "test@test.com",
		Name:  "testname",
	}
	secret := "somesecret"
	dtl := 7 // days

	t.Run("Successfully sign and decode a token", func(t *testing.T) {
		tkn := Sign(c, []byte(secret), dtl)

		data, err := Decode(tkn, []byte(secret))

		assert.Nil(t, err, "Error decoding token")
		assert.Equal(t, data["email"], "test@test.com")
		assert.Equal(t, data["sub"], "testname")
	})

	t.Run("Fail to decode a tampered token", func(t *testing.T) {
		tkn := Sign(c, []byte(secret), dtl)
		tamperedToken := tkn + "hehehe"

		_, err := Decode(tamperedToken, []byte(secret))
		assert.NotNil(t, err, "There should be an error decoding")
	})

	t.Run("Fail to validate an expired token", func(t *testing.T) {
		tkn := Sign(c, []byte(secret), -5) // 5 days in the past from now

		_, err := Decode(tkn, []byte(secret))

		assert.NotNil(t, err, "Should be an error decoding a token")
	})
}
