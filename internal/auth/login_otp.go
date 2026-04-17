package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func generateLoginOTP() (string, error) {
	max := 1000000
	n, err := rand.Int(rand.Reader, bigInt(int64(max)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashLoginOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
