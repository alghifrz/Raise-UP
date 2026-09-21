package bot

import (
	"strings"
	"unicode"
)

// Intent is the high-level classification of an inbound citizen message.
type Intent string

const (
	IntentWelcome   Intent = "welcome"
	IntentMenu      Intent = "menu"
	IntentComplaint Intent = "complaint"
	IntentOther     Intent = "other"
)

// Session is lightweight conversation context used for classification.
type Session struct {
	DisplayName    string
	IsFirstMessage bool
}

// Decision is the rule-based outcome for one inbound message.
type Decision struct {
	Intent        Intent
	ShouldWelcome bool
	MenuChoice    string // "1".."4", empty when not a menu pick
	Reply         string
}

// Classify maps free-text WhatsApp input to a Decision (no LLM).
func Classify(text string, sess Session) Decision {
	raw := strings.TrimSpace(text)
	normalized := normalizeText(raw)
	name := displayName(sess.DisplayName)

	if choice, ok := menuChoice(normalized); ok {
		return Decision{
			Intent:     IntentMenu,
			MenuChoice: choice,
			Reply:      FormatMenuAck(choice),
		}
	}

	if isMenuRequest(normalized) {
		return Decision{
			Intent:        IntentWelcome,
			ShouldWelcome: true,
			Reply:         FormatWelcome(name),
		}
	}

	if looksLikeComplaint(normalized) {
		return Decision{
			Intent: IntentComplaint,
			Reply:  FormatAskComplaint(),
		}
	}

	if sess.IsFirstMessage || isGreeting(normalized) {
		return Decision{
			Intent:        IntentWelcome,
			ShouldWelcome: true,
			Reply:         FormatWelcome(name),
		}
	}

	return Decision{
		Intent: IntentOther,
		Reply:  FormatMenuHint(),
	}
}

func displayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Warga"
	}
	return name
}

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func menuChoice(normalized string) (string, bool) {
	switch normalized {
	case "1", "2", "3", "4":
		return normalized, true
	default:
		return "", false
	}
}

func isMenuRequest(normalized string) bool {
	switch normalized {
	case "menu", "bantuan", "help", "pilihan", "mulai", "start":
		return true
	}
	return false
}

func isGreeting(normalized string) bool {
	greetings := []string{
		"halo", "hallo", "hai", "hi", "hello",
		"selamat pagi", "selamat siang", "selamat sore", "selamat malam",
		"assalamualaikum", "assalamu'alaikum", "asalamualaikum",
		"permisi", "ping",
	}
	for _, g := range greetings {
		if normalized == g || strings.HasPrefix(normalized, g+" ") {
			return true
		}
	}
	return false
}

func looksLikeComplaint(normalized string) bool {
	keywords := []string{
		"pengaduan", "keluhan", "lapor", "laporan",
		"sampah", "banjir", "bocor", "macet", "rusak",
		"kebisingan", "keamanan", "maling", "pencurian",
		"listrik", "air mati", "got", "selokan", "jalan berlubang",
	}
	for _, kw := range keywords {
		if strings.Contains(normalized, kw) {
			return true
		}
	}
	return false
}
