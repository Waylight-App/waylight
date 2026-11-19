package pushproto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// version is the first byte of a sealed payload.
const version = 0x01

// info binds the derived key to this protocol and version.
const info = "waylight-push-v1"

// MaxPayload is the largest sealed payload the relay forwards; FCM and APNs
// accept 4 KB including their own fields.
const MaxPayload = 3000

// ErrPayload means a sealed payload cannot be opened.
var ErrPayload = errors.New("pushproto: cannot open payload")

func derive(shared []byte) (cipher.AEAD, error) {
	key, err := hkdf.Key(sha256.New, shared, nil, info, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// Seal encrypts plaintext for the device's X25519 public key: an ephemeral
// key agreement, HKDF-SHA256 and AES-256-GCM. The result is
// version ‖ ephemeral public key ‖ nonce ‖ ciphertext.
func Seal(device *ecdh.PublicKey, plaintext []byte) ([]byte, error) {
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	shared, err := eph.ECDH(device)
	if err != nil {
		return nil, err
	}
	aead, err := derive(shared)
	if err != nil {
		return nil, err
	}
	out := append([]byte{version}, eph.PublicKey().Bytes()...)
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	out = append(out, nonce...)
	return aead.Seal(out, nonce, plaintext, nil), nil
}

// Open decrypts a sealed payload with the device's private key. The app
// implements the same steps; Open documents them and serves the tests.
func Open(device *ecdh.PrivateKey, sealed []byte) ([]byte, error) {
	const head = 1 + 32
	if len(sealed) < head+12 || sealed[0] != version {
		return nil, ErrPayload
	}
	eph, err := ecdh.X25519().NewPublicKey(sealed[1:head])
	if err != nil {
		return nil, ErrPayload
	}
	shared, err := device.ECDH(eph)
	if err != nil {
		return nil, ErrPayload
	}
	aead, err := derive(shared)
	if err != nil {
		return nil, err
	}
	nonce := sealed[head : head+aead.NonceSize()]
	plain, err := aead.Open(nil, nonce, sealed[head+aead.NonceSize():], nil)
	if err != nil {
		return nil, ErrPayload
	}
	return plain, nil
}
