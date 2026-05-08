package cache

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const defaultTTLDays = 7

func dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "web-research")
}

func key(url string) string {
	return fmt.Sprintf("%x.md", sha256.Sum256([]byte(url)))
}

func ttl() time.Duration {
	days := defaultTTLDays
	if d := os.Getenv("WR_CACHE_DAYS"); d != "" {
		if n, err := strconv.Atoi(d); err == nil {
			days = n
		}
	}
	return time.Duration(days) * 24 * time.Hour
}

func Get(url string) (string, bool) {
	path := filepath.Join(dir(), key(url))
	info, err := os.Stat(path)
	if err != nil || time.Since(info.ModTime()) > ttl() {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func Set(url, content string) error {
	if err := os.MkdirAll(dir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir(), key(url)), []byte(content), 0644)
}
