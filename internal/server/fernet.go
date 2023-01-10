package server

import (
	"errors"

	"github.com/fernet/fernet-go"
)

// Encrypt takes the plaintext secret, generates a key and produces a
// token.
func Encrypt(secret string) (token, key string, err error) {
	fernetkey := fernet.Key{}
	err = fernetkey.Generate()
	if err != nil {
		return "", "", err
	}

	tokenBytes, err := fernet.EncryptAndSign([]byte(secret), &fernetkey)
	if err != nil {
		return "", "", err
	}

	return string(tokenBytes), fernetkey.Encode(), nil
}

// Decrypt decodes key and decrypts token with it.
func Decrypt(token string, key string) (string, error) {
	fernetkey, err := fernet.DecodeKey(key)
	if err != nil {
		return "", err
	}

	msg := fernet.VerifyAndDecrypt([]byte(token), 0, []*fernet.Key{fernetkey})
	if len(msg) == 0 {
		return "", errors.New("secret was empty")
	}

	return string(msg), nil
}
