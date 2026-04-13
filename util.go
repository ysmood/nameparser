package nameparser

import (
	"fmt"
	"strings"
	"unicode"
)

// lc lowercases and strips trailing periods, matching python-nameparser behavior.
func lc(value string) string {
	if value == "" {
		return ""
	}
	return strings.Trim(strings.ToLower(value), ".")
}

func copyStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}

func indexOf(items []string, value string, start int) int {
	if start < 0 {
		start = 0
	}
	for i := start; i < len(items); i++ {
		if items[i] == value {
			return i
		}
	}
	return -1
}

func firstMatch(items []string, fn func(string) bool) (string, bool) {
	for _, item := range items {
		if fn(item) {
			return item, true
		}
	}
	return "", false
}

func joinSpace(items []string) string {
	return strings.Join(items, " ")
}

func titleCaseWord(word string) string {
	runes := []rune(strings.ToLower(word))
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func sanitizeFormatOutput(s string, emptyDefault string) string {
	if emptyDefault != "" {
		s = strings.ReplaceAll(s, emptyDefault, "")
	}
	s = strings.ReplaceAll(s, " ()", "")
	s = strings.ReplaceAll(s, " ''", "")
	s = strings.ReplaceAll(s, " \"\"", "")
	return s
}

func normalizeEncodingName(enc string) string {
	enc = strings.ToLower(strings.TrimSpace(enc))
	switch enc {
	case "latin-1", "latin1", "latin_1", "iso-8859-1", "iso8859-1":
		return "latin_1"
	case "utf-8", "utf8", "":
		return "utf-8"
	default:
		return enc
	}
}

func decodeBytes(input []byte, encoding string) (string, error) {
	switch normalizeEncodingName(encoding) {
	case "utf-8":
		return string(input), nil
	case "latin_1":
		runes := make([]rune, len(input))
		for i, b := range input {
			runes[i] = rune(b)
		}
		return string(runes), nil
	default:
		return "", fmt.Errorf("unsupported encoding: %s", encoding)
	}
}
