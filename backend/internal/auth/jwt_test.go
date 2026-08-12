package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "securepassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, CheckPasswordHash(password, hash))
	assert.False(t, CheckPasswordHash("wrongpassword", hash))
}

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-key"
	token, err := GenerateToken("user-123", "test@example.com", "Test User", secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseToken(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "Test User", claims.Name)
}

func TestParseTokenInvalidSecret(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"
	token, err := GenerateToken("user-123", "test@example.com", "Test User", secret)
	assert.NoError(t, err)

	_, err = ParseToken(token, wrongSecret)
	assert.Error(t, err)
}

func TestParseTokenEmptyToken(t *testing.T) {
	_, err := ParseToken("", "secret")
	assert.Error(t, err)
}

func TestParseTokenMalformedToken(t *testing.T) {
	_, err := ParseToken("not.a.token", "secret")
	assert.Error(t, err)
}