package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

//go:generate mockgen -source ./problem.go -destination ../../mock/service/mock_config_service.go -package mock

type AesService interface {
	AESEncrypt(plaintext []byte) (string, error)
	AESDecrypt(ciphertext string) ([]byte, error)
	pkcs7Pad(data []byte, blockSize int) []byte
	pkcs7Unpad(data []byte, blockSize int) ([]byte, error)
}
type aesService struct {
	Key []byte
}

// AESEncrypt 使用 AES-256-CBC 模式加密数据
func (s *aesService) AESEncrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(s.Key)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	plaintext = s.pkcs7Pad(plaintext, block.BlockSize())

	// 创建 CBC 模式
	ciphertext := make([]byte, len(plaintext))
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(ciphertext, plaintext)

	// 将 IV 和密文拼接并 base64 编码
	result := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// AESDecrypt 使用 AES-256-CBC 模式解密数据
func (s *aesService) AESDecrypt(ciphertext string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(s.Key)
	if err != nil {
		return nil, err
	}

	// 分离 IV 和密文
	ivSize := block.BlockSize()
	if len(decoded) < ivSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := decoded[:ivSize]
	ciphertextBytes := decoded[ivSize:]

	// 创建 CBC 解密器
	stream := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertextBytes))
	stream.CryptBlocks(plaintext, ciphertextBytes)

	// 移除 PKCS7 padding
	return s.pkcs7Unpad(plaintext, block.BlockSize())
}

// pkcs7Pad 对数据进行 PKCS7 填充
func (s *aesService) pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpad 移除 PKCS7 填充
func (s *aesService) pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > blockSize {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}
