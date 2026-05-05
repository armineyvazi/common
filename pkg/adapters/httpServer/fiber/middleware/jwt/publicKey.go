package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const publicKeyFileName = "public.pem"

func getPublicKey(publicKey string) (*rsa.PublicKey, error) {
	err := writeKeytoFile(publicKey)
	if err != nil {
		return nil, err
	}

	key, err := loadPublicKey()
	if err != nil {
		return nil, err
	}

	if key == nil {
		return nil, fmt.Errorf("could not load public key")
	}

	return key, nil
}

func writeKeytoFile(key string) error {
	err := os.WriteFile(publicKeyFileName, []byte(key), 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadPublicKey() (*rsa.PublicKey, error) {
	file, err := os.Open(publicKeyFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	keyBytes := make([]byte, fileSize)

	_, err = file.Read(keyBytes)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(keyBytes))
	if block == nil {
		return nil, err
	}

	if block.Type != "RSA PUBLIC KEY" {
		return nil, err
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}
