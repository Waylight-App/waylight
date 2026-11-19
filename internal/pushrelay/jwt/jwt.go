// Package jwt signs the compact JSON Web Tokens that FCM (RS256) and APNs
// (ES256) need for authentication. Only signing is implemented; the relay
// never verifies tokens.
package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
)

var b64 = base64.RawURLEncoding

// Sign returns header.claims.signature. The key decides the algorithm:
// RS256 for RSA, ES256 (raw r‖s, RFC 7518 §3.4) for P-256.
func Sign(key crypto.Signer, header, claims map[string]any) (string, error) {
	switch key.(type) {
	case *rsa.PrivateKey:
		header["alg"] = "RS256"
	case *ecdsa.PrivateKey:
		header["alg"] = "ES256"
	default:
		return "", errors.New("jwt: unsupported key")
	}
	header["typ"] = "JWT"
	h, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	c, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	input := b64.EncodeToString(h) + "." + b64.EncodeToString(c)
	sum := sha256.Sum256([]byte(input))
	var sig []byte
	switch k := key.(type) {
	case *rsa.PrivateKey:
		sig, err = rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, sum[:])
	case *ecdsa.PrivateKey:
		r, s, e := ecdsa.Sign(rand.Reader, k, sum[:])
		if e != nil {
			return "", e
		}
		sig = make([]byte, 64)
		r.FillBytes(sig[:32])
		s.FillBytes(sig[32:])
	}
	if err != nil {
		return "", err
	}
	return input + "." + b64.EncodeToString(sig), nil
}

// ParseKey reads a PKCS#8 or PKCS#1 private key in PEM, as Google's service
// account files and Apple's .p8 keys contain.
func ParseKey(pemData []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("jwt: no PEM block")
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if s, ok := k.(crypto.Signer); ok {
			return s, nil
		}
		return nil, fmt.Errorf("jwt: unsupported key %T", k)
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	return nil, errors.New("jwt: cannot parse private key")
}
