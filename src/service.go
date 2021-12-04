package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

func getID() string {
	h := sha256.New()
	h.Write([]byte(strconv.Itoa(time.Now().Nanosecond())))
	return hex.EncodeToString(h.Sum(nil))
}
