package security

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

// GetID return sha256 value
func GetID() string {
	h := sha256.New()
	h.Write([]byte(strconv.Itoa(time.Now().Nanosecond())))
	return hex.EncodeToString(h.Sum(nil))
}
