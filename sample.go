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

func RandStringWithCharset(length int, charset string) string {
	// for English letters, we can just use bytes instead
	// of runes.
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return RandStringWithCharset(length, charset)
}

func GetRandUsername() string {
	return fmt.Sprintf("user_%s%s", RandStringWithCharset(3, letters), RandStringWithCharset(3, digits))
}

func GetRandPassword() string {
	return RandString(10)
}

func GetRandChannelName() string {
	return fmt.Sprintf("chan_%s%s", RandStringWithCharset(3, letters), RandStringWithCharset(3, digits))
}

func GetRandBool() bool {
	return seededRand.Intn(2) == 1
}
