package whatsapp

import (
	"regexp"
	"strings"
)

var nonDigit = regexp.MustCompile(`\D+`)

// NormalizePhone converts common ID phone formats to WhatsApp digits (no +).
// Examples: "0812..." -> "62812...", "+62 812-..." -> "62812..."
func NormalizePhone(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidRequest
	}

	digits := nonDigit.ReplaceAllString(trimmed, "")
	if digits == "" {
		return "", ErrInvalidRequest
	}

	switch {
	case strings.HasPrefix(digits, "62"):
		// already country-coded
	case strings.HasPrefix(digits, "0"):
		digits = "62" + digits[1:]
	case len(digits) >= 9 && len(digits) <= 12 && digits[0] == '8':
		// local mobile without leading 0
		digits = "62" + digits
	}

	if len(digits) < 10 || len(digits) > 15 {
		return "", ErrInvalidRequest
	}
	return digits, nil
}

// DigitsOnly strips non-digits (for comparing resident phones).
func DigitsOnly(raw string) string {
	return nonDigit.ReplaceAllString(raw, "")
}

// PhoneMatchCandidates returns variants useful for matching stored resident phones.
func PhoneMatchCandidates(e164Digits string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	add := func(v string) {
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	add(e164Digits)
	add("+" + e164Digits)
	if strings.HasPrefix(e164Digits, "62") && len(e164Digits) > 2 {
		local := "0" + e164Digits[2:]
		add(local)
		add(e164Digits[2:])
	}
	return out
}
