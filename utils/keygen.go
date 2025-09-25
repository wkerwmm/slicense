package utils

import (
	"math/rand"
	"strings"
	"time"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateLicenseKey() string {
	rand.Seed(time.Now().UnixNano())
	var sb strings.Builder

	for i := 0; i < 4; i++ {
		if i > 0 {
			sb.WriteString("-")
		}
		for j := 0; j < 4; j++ {
			sb.WriteByte(chars[rand.Intn(len(chars))])
		}
	}

	return sb.String()
}