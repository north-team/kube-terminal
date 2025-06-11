package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const SALTED_STR = "Salted__"
const SECRET_PASS = "Fit2cloud@2015"

func Decrypt(password string, source string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return "", err
	}
	if !bytes.HasPrefix(data, []byte(SALTED_STR)) {
		return "", errors.New("input is not salted")
	}
	salt := data[8:16] // skip 'Salted__'
	key, iv := evpBytesToKey([]byte(password), salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	cleartext := make([]byte, len(data)-16)
	mode.CryptBlocks(cleartext, data[16:])
	cleartext = pkcs5Unpad(cleartext)
	return string(cleartext), nil
}

func Encrypt(password string, clearText string) (string, error) {
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	key, iv := evpBytesToKey([]byte(password), salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plain := pkcs5Pad([]byte(clearText), aes.BlockSize)
	ciphertext := make([]byte, len(plain))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plain)

	final := append([]byte(SALTED_STR), salt...)
	final = append(final, ciphertext...)
	return base64.StdEncoding.EncodeToString(final), nil
}

func evpBytesToKey(password []byte, salt []byte) ([]byte, []byte) {
	hasher := md5.New()
	keyIv := []byte{}
	hash := []byte{}
	for len(keyIv) < 32 {
		hasher.Reset()
		hasher.Write(hash)
		hasher.Write(password)
		hasher.Write(salt)
		hash = hasher.Sum(nil)
		keyIv = append(keyIv, hash...)
	}
	return keyIv[:16], keyIv[16:32]
}

func pkcs5Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(data, padText...)
}

func pkcs5Unpad(data []byte) []byte {
	length := len(data)
	padLen := int(data[length-1])
	return data[:(length - padLen)]
}
