package checksum

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func GenerateChecksum(i interface{}) (string, error) {
	// bytes, err := json.Marshal(object)
	// if err != nil {
	// 	return "", err
	// }

	hash := sha256.New()
	hash.Write(i)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	return checksum, nil
}
