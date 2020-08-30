// File for the generation of random objects, suitable for testing needs
package accord

import (
	"fmt"
	"math/rand"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digits = "0123456789"
const charset = letters + digits

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()),
)

func StringWithCharset(length int, charset string) string {
	// for English letters, we can just use bytes instead
	// of runes.
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func GetRandUsername() string {
	return fmt.Sprintf("user_%s%s", StringWithCharset(3, letters), StringWithCharset(3, digits))
}

func GetRandPassword() string {
	return String(10)
}

func GetRandChannelName() string {
	return fmt.Sprintf("chan_%s%s", StringWithCharset(3, letters), StringWithCharset(3, digits))
}

func GetRandBool() bool {
	return seededRand.Intn(2) == 1
}
