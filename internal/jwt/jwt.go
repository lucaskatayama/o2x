package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	Picture       string `json:"picture,omitempty"`
	Scope         string `json:"scope,omitempty"`
}

type Validator struct {
	jwksURI string
	keys    map[string]interface{}
	mu      sync.RWMutex
	httpCli *http.Client
}

func NewValidator(jwksURI string) *Validator {
	return &Validator{
		jwksURI: jwksURI,
		keys:    make(map[string]interface{}),
		httpCli: &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *Validator) Decode(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	payload, err := base64Decode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}

	return &claims, nil
}

func (v *Validator) Validate(tokenString string) (*Claims, error) {
	claims, err := v.Decode(tokenString)
	if err != nil {
		return nil, err
	}

	if v.jwksURI == "" {
		return claims, nil
	}

	if err := v.fetchJWKS(); err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}
		v.mu.RLock()
		key, ok := v.keys[kid]
		v.mu.RUnlock()
		if !ok {
			return nil, fmt.Errorf("key %s not found in JWKS", kid)
		}
		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

func (v *Validator) fetchJWKS() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	resp, err := v.httpCli.Get(v.jwksURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
			Use string `json:"use"`
			Alg string `json:"alg"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			pubKey, err := parseRSAPublicKey(key.N, key.E)
			if err != nil {
				continue
			}
			v.keys[key.Kid] = pubKey
		}
	}

	return nil
}

func parseRSAPublicKey(nStr, eStr string) (interface{}, error) {
	nBytes, err := base64Decode(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64Decode(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &struct {
		N *big.Int
		E int
	}{N: n, E: e}, nil
}

func base64Decode(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
