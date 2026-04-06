package shared

import (
	"strings"
	"time"
)

func NormalizeHash(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
