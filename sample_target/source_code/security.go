package security

import (
	"crypto/des"
	"crypto/md5"
	"crypto/rsa"
	"crypto/rand"
	"encoding/hex"
)

// LegacyHash performs MD5 hash computation
func LegacyHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// LegacyEncrypt performs weak DES encryption
func LegacyEncrypt(key, text []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(text))
	block.Encrypt(ciphertext, text)
	return ciphertext, nil
}

// LegacyKeyGen generates RSA 1024-bit key pair
func LegacyKeyGen() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 1024)
}
