package hash

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/bcrypt"
)

func Hash(v string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Compare(v string, i string) error {
	return bcrypt.CompareHashAndPassword([]byte(v), []byte(i))
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func EncAES(v string, key string) (string, error) {
	block, err := aes.NewCipher([]byte(createHash(key)))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(v), nil)), nil
}

func CompareAES(hash string, key string, v string) bool {
	decrypted, err := DecAES(hash, key)
	if err != nil {
		return false
	}
	return decrypted == v
}

func DecAES(v string, key string) (string, error) {
	vByte, err := hex.DecodeString(v)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(createHash(key)))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := vByte[:nonceSize], vByte[nonceSize:]
	res, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
