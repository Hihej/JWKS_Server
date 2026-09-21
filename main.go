package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"
	"encoding/base64"
	"encoding/json"
	"math/big"

	"github.com/golang-jwt/jwt/v5"
)

type Key struct {
	// Key stores an RSA private key along with its ID and expiration time.
	PrivateKey *rsa.PrivateKey
	Kid        string
	ExpiresAt  time.Time
}

var validKey Key
var expiredKey Key
// generateKeys creates one valid RSA key and one expired RSA key.
func generateKeys() {
func generateKeys() {
	// Current key
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	// Generate a key that will be valid for one hour.
	validKey = Key{
		PrivateKey: privateKey,
		Kid:        "valid-key",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	// Expired key
	oldPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	// Generate a second key that expired one hour ago.
	expiredKey = Key{
		PrivateKey: oldPrivateKey,
		Kid:        "expired-key",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
}
// jwksHandler returns the valid public key in JWKS format.
func jwksHandler(w http.ResponseWriter, r *http.Request) {
	publicKey := validKey.PrivateKey.PublicKey

	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())

	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]string{
		"kty": "RSA",
		"use": "sig",
		"kid": validKey.Kid,
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}

	response := map[string]interface{}{
		"keys": []interface{}{jwk},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := validKey
	expiration := time.Now().Add(1 * time.Hour)

	if r.URL.Query().Get("expired") == "true" {
		key = expiredKey
		expiration = time.Now().Add(-1 * time.Hour)
	}

	claims := jwt.MapClaims{
		"sub": "user",
		"iat": time.Now().Unix(),
		"exp": expiration.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	token.Header["kid"] = key.Kid

	signedToken, err := token.SignedString(key.PrivateKey)

	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/jwt")
	w.Write([]byte(signedToken))
}
func setupServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/.well-known/jwks.json", jwksHandler)
	mux.HandleFunc("/auth", authHandler)

	return mux
}

func main() {
	generateKeys()

	fmt.Println("Valid key ID:", validKey.Kid)
	fmt.Println("Valid key expires:", validKey.ExpiresAt)
	fmt.Println("Expired key ID:", expiredKey.Kid)
	fmt.Println("Expired key expired:", expiredKey.ExpiresAt)
	fmt.Println("JWKS Server running on port 8080")

	http.ListenAndServe(":8080", setupServer())
}