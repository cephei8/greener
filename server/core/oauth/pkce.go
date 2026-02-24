package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

func ValidatePKCE(codeVerifier, codeChallenge, method string) bool {
	if method != "S256" {
		return false
	}

	if codeVerifier == "" || codeChallenge == "" {
		return false
	}

	h := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(codeChallenge)) == 1
}

func GenerateCodeChallenge(codeVerifier string) string {
	h := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
