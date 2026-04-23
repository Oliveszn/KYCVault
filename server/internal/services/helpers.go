package services

import (
	"crypto/sha256"
	"fmt"
)

func computeChecksum(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
