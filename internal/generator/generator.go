package generator

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var counter int64 // 原子计数器，保证并发安全

// Generate 生成 7 位短码
// 策略：MD5(url + timestamp + counter) -> base62 取前7位
// 即使相同 URL 每次生成也不同（因为 timestamp + counter）
func Generate(url string) string {
	ts := time.Now().UnixNano()
	seq := atomic.AddInt64(&counter, 1)
	input := fmt.Sprintf("%s:%d:%d", url, ts, seq)

	hash := md5.Sum([]byte(input))
	hexStr := hex.EncodeToString(hash[:])

	// hex -> base62
	code := hexToBase62(hexStr)
	if len(code) > 7 {
		code = code[:7]
	}
	return code
}

func hexToBase62(hex string) string {
	var result string
	var val int64

	for _, c := range hex {
		switch {
		case c >= '0' && c <= '9':
			val = val*16 + int64(c-'0')
		case c >= 'a' && c <= 'f':
			val = val*16 + int64(c-'a'+10)
		case c >= 'A' && c <= 'F':
			val = val*16 + int64(c-'A'+10)
		}
	}

	if val == 0 {
		return "0"
	}

	for val > 0 {
		result = string(base62Chars[val%62]) + result
		val /= 62
	}

	return result
}
