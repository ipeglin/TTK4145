package checksum

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func GenerateChecksum(i interface{}) (string, error) {
	hash := sha256.New()
	hash.Write(i)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	return checksum, nil
}
