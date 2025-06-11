package aes

import (
	"fmt"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	ciphertext := "Password123@influxdb"
	encrypt, err := Encrypt(SECRET_PASS, ciphertext)
	fmt.Println(encrypt)
	//ciphertext := "U2FsdGVkX1+LiEy6OLJPir1WquBpArny9/0l4njPEs+vi9tGIDieYh+zfT/nQCT2"
	decrypt, err := Decrypt(SECRET_PASS, encrypt)
	if err != nil {
		panic(err)
	}
	fmt.Println(decrypt)
}
