package gophcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Cryptographer struct {
	cipher cipher.AEAD
	vector []byte
}

func New(cryptoKey []byte) (*Cryptographer, error) {
	aesblock, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	s := sha256.Sum256(cryptoKey)
	vec := s[:aesgcm.NonceSize()]

	return &Cryptographer{cipher: aesgcm, vector: vec}, nil
}

func (c *Cryptographer) Encrypt(data []byte) string {
	dst := c.cipher.Seal(nil, c.vector, data, nil)

	return base64.RawStdEncoding.EncodeToString(dst)
}

func (c *Cryptographer) Decrypt(base64data []byte) ([]byte, error) {
	data := make([]byte, base64.RawStdEncoding.DecodedLen(len(base64data)))

	_, err := base64.RawStdEncoding.Decode(data, base64data)
	if err != nil {
		return nil, fmt.Errorf("base64 decode err=%w", err)
	}

	dst, err := c.cipher.Open(nil, c.vector, data, nil)
	if err != nil {
		return nil, fmt.Errorf("decode err=%w", err)
	}

	return dst, nil
}

func GenerateCryptoKey() ([]byte, error) {
	const keyLenDefault = aes.BlockSize

	data, err := generateRandom(keyLenDefault)
	if err != nil {
		return nil, nil
	}

	return data, nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
