package secret

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSecretKey() (string, error) {
	b := make([]byte, 32)

	// 使用加密级随机数
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	token := hex.EncodeToString(b)
	return token, nil
}
