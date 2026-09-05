package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	cryptRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"log/slog"
)

const aesKeySize = 32

func ParsePublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse key error: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}
	var priv *rsa.PrivateKey
	if block.Type == "RSA PRIVATE KEY" {
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else {
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse key error: %w (pkcs1: %v)", err2, err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		priv = rsaKey
	}
	if err != nil {
		return nil, fmt.Errorf("parse key error: %w", err)
	}
	return priv, nil
}

func Encrypt(pubKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	aesKey := make([]byte, aesKeySize)
	if _, err := cryptRand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("generate AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err2 := cryptRand.Read(nonce); err2 != nil {
		return nil, fmt.Errorf("generate nonce: %w", err2)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, []byte{})

	encAESKey, err := rsa.EncryptOAEP(sha256.New(), cryptRand.Reader, pubKey, aesKey, []byte{})
	if err != nil {
		return nil, fmt.Errorf("encrypt AES key with RSA: %w", err)
	}

	keyLen := make([]byte, 2)
	binary.BigEndian.PutUint16(keyLen, uint16(len(encAESKey)))

	result := make([]byte, 0, 2+len(encAESKey)+len(ciphertext))
	result = append(result, keyLen...)
	result = append(result, encAESKey...)
	result = append(result, ciphertext...)

	return result, nil
}

func Decrypt(privKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("data too short")
	}
	keyLen := binary.BigEndian.Uint16(data[:2])
	data = data[2:]

	if len(data) < int(keyLen) {
		return nil, fmt.Errorf("encrypted key data too short")
	}
	encAESKey := data[:keyLen]
	ciphertext := data[keyLen:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), nil, privKey, encAESKey, []byte{})
	if err != nil {
		return nil, fmt.Errorf("decrypt AES key with RSA: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	plaintext, err := gcm.Open([]byte{}, nonce, ciphertext, []byte{})
	if err != nil {
		return nil, fmt.Errorf("decrypt with AES-GCM: %w", err)
	}
	return plaintext, nil
}

func Middleware(privKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/octet-stream" || privKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				slog.Error("crypto middleware: read body", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			decrypted, err := Decrypt(privKey, body)
			if err != nil {
				slog.Error("crypto middleware: decrypt", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.Header.Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}
