package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	jwksCache struct {
		sync.RWMutex
		keys   map[string]interface{}
		expiry time.Time
	}
	// jwksURL = "https://" + os.Getenv("AUTH0_DOMAIN") + "/.well-known/jwks.json"
)

func fetchJWKS() error {

	// Adapt below to AWS, ideally this is all that is needed
	// jwksURL := "https://cognito-idp." + os.Getenv("AWS_REGION") + ".amazonaws.com/" + os.Getenv("COGNITO_USER_POOL_ID") + "/.well-known/jwks.json"

	jwksURL := "https://" + os.Getenv("AUTH0_DOMAIN") + "/.well-known/jwks.json"

	resp, err := http.Get(jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var jwks struct {
		Keys []struct {
			Kty string   `json:"kty"`
			Kid string   `json:"kid"`
			Use string   `json:"use"`
			N   string   `json:"n"`
			E   string   `json:"e"`
			X5c []string `json:"x5c"`
		} `json:"keys"`
	}

	if err := json.Unmarshal(body, &jwks); err != nil {
		return err
	}

	newKeys := make(map[string]interface{})
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" {
			parsedKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte("-----BEGIN CERTIFICATE-----\n" + key.X5c[0] + "\n-----END CERTIFICATE-----"))
			if err != nil {
				continue // or log the error
			}
			newKeys[key.Kid] = parsedKey
		}
	}

	jwksCache.Lock()
	jwksCache.keys = newKeys
	jwksCache.expiry = time.Now().Add(24 * time.Hour) // Cache for 24 hours
	jwksCache.Unlock()

	return nil
}

func getKeyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// Ensure the token algorithm matches your expectations
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		jwksCache.RLock()
		key, ok := jwksCache.keys[kid]
		jwksCache.RUnlock()
		if ok && key != nil {
			return key, nil
		}

		// Refresh JWKS if the key is not found (may handle differently based on your needs)
		if err := fetchJWKS(); err != nil {
			return nil, err
		}

		jwksCache.RLock()
		key, ok = jwksCache.keys[kid]
		jwksCache.RUnlock()
		if ok && key != nil {
			return key, nil
		}

		return nil, fmt.Errorf("key %v not found in JWKS", kid)
	}
}
