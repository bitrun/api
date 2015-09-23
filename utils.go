package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
)

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func expandPath(path string) string {
	if strings.Index(path, "~") == 0 {
		path = strings.Replace(path, "~", os.Getenv("HOME"), 1)
	}

	return path
}
