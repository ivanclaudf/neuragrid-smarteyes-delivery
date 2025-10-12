package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptAES256 encrypts data using AES-256
func EncryptAES256(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// DecryptAES256 decrypts data using AES-256
func DecryptAES256(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("decryption key must be 32 bytes for AES-256")
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// EncryptAndBase64Encode encrypts the data with AES-256 and encodes it as base64
func EncryptAndBase64Encode(data []byte, key []byte) (string, error) {
	encrypted, err := EncryptAES256(data, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// EncryptAndEncodeBase64 encrypts a string with AES-256 and encodes it as base64
func EncryptAndEncodeBase64(plaintext string, key []byte) (string, error) {
	return EncryptAndBase64Encode([]byte(plaintext), key)
}

// DecodeBase64AndDecrypt decodes base64 data and decrypts it with AES-256
func DecodeBase64AndDecrypt(encoded string, key []byte) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return DecryptAES256(decoded, key)
}

// Base64Encode encodes a byte array to base64 string
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
