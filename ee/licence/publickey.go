package licence

import (
	"crypto/ed25519"
	"encoding/base64"
)

// publicKey checks the signature of licence keys. The matching private key
// and the signing tool are kept outside this repository (#307); replacing
// this value makes every key issued so far invalid.
var publicKey = mustKey("8bd1kdfUVryRRLlKMgWEJxKqWUXoJKZ+JrT/rM5GUVA=")

func mustKey(b64 string) ed25519.PublicKey {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(raw) != ed25519.PublicKeySize {
		panic("ee/licence: invalid public key")
	}
	return raw
}
