package checksum

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func GenerateJSONChecksum(object interface{}) (string, error) {
	bytes, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	// Calculate SHA256 checksum
	hash := sha256.New()
	hash.Write(bytes)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	return checksum, nil
}