package cryptomap

import (
	"crypto"
	"log"
	"strings"

	// Importing all the currently supported hashes
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
)

// DefaultCipher is the crypto cipher default used if none is specified or
// specified is unknown.
var DefaultCipher = "sha1"

// Ciphers is the known set of mappings for string to crypto.Hash
// use an init() to add custom hash ciphers
var Ciphers = map[string]crypto.Hash{
	"md5":    crypto.MD5,
	"sha1":   crypto.SHA1,
	"sha224": crypto.SHA224,
	"sha256": crypto.SHA256,
	"sha384": crypto.SHA384,
	"sha512": crypto.SHA512,
}

// DetermineHash takes a generic string, like "sha1" and returns the
// corresponding crypto.Hash
func DetermineHash(str string) (h crypto.Hash) {
	if h, ok := Ciphers[strings.ToLower(str)]; ok {
		return h
	}
	log.Printf("WARNING: unknown cipher %q. using %q", str, DefaultCipher)
	return Ciphers[DefaultCipher]
}
