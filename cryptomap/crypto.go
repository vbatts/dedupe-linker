package cryptomap

import (
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"log"
	"strings"
)

var knownCiphers = map[string]crypto.Hash{
	"md5": crypto.MD5,
}

func DetermineHash(str string) (h crypto.Hash) {
	switch strings.ToLower(str) {
	case "md5":
		h = crypto.MD5
	case "sha1":
		h = crypto.SHA1
	case "sha224":
		h = crypto.SHA224
	case "sha256":
		h = crypto.SHA256
	case "sha384":
		h = crypto.SHA384
	case "sha512":
		h = crypto.SHA512
	default:
		log.Println("WARNING: unknown cipher %q. using 'sha1'", str)
		h = crypto.SHA1
	}

	return h
}
