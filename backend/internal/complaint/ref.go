package complaint

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const maxRefAttempts = 5

// GenerateRef builds a human-readable complaint reference: CMP-YYYYMMDD-XXXX.
func GenerateRef(now time.Time) (string, error) {
	var buf [2]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate ref suffix: %w", err)
	}
	suffix := strings.ToUpper(hex.EncodeToString(buf[:]))
	return fmt.Sprintf("CMP-%s-%s", now.UTC().Format("20060102"), suffix), nil
}
