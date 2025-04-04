package datasource

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GenerateKeyBase64(length int) string {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Erro ao gerar chave: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

// chave 32 bytes (AES-256)
var encryptionKey []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}
	keyStr := os.Getenv("ENCRYPTION_KEY")
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		log.Fatal("Erro ao decodificar a chave:", err)
	}
	encryptionKey = keyBytes
}

func Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("nonce inválido")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
