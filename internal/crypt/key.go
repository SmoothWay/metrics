// Package crypt responsible for encrypting and decrypting requests and responses
package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

func ReadKeyFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecodeKey(data)
}

func DecodeKey(data []byte) ([]byte, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode key file")
	}
	return block.Bytes, nil
}

func Encrypt(data []byte, pubKey []byte) ([]byte, error) {
	key, err := x509.ParsePKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	data, err = EncryptChunks(rand.Reader, key.(*rsa.PublicKey), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// EncryptChunks encrypt the message in chunks if the message is larger than the key length
func EncryptChunks(random io.Reader, pub *rsa.PublicKey, msg []byte) ([]byte, error) {
	msgLen := len(msg)
	step := pub.Size() - 11
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := rsa.EncryptPKCS1v15(random, pub, msg[start:finish])
		if err != nil {
			return nil, err
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

func Decrypt(data []byte, privateKey []byte) ([]byte, error) {
	key, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	data, err = DecryptChunks(nil, key, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DecryptChunks decrypt the message in chunks if the message is larger than the key length
func DecryptChunks(random io.Reader, priv *rsa.PrivateKey, msg []byte) ([]byte, error) {
	msgLen := len(msg)
	step := priv.PublicKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptPKCS1v15(random, priv, msg[start:finish])
		if err != nil {
			return nil, err
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}
