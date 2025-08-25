package str

import (
	"slices"
	"strings"
)

// BreakLine breaks a line at word boundaries if it exceeds the limit.
func BreakLine(line string, limit int) string {
	if len(line) <= limit {
		return line
	}

	words := strings.Fields(line)
	var out strings.Builder

	var lineLen int
	for word := range slices.Values(words) {
		wordlen := len(word)
		if lineLen == 0 {
			out.WriteString(word)
			lineLen += wordlen
		} else if lineLen+wordlen < limit {
			out.WriteByte(' ') // add space
			out.WriteString(word)
			lineLen += wordlen + 1
		} else {
			out.WriteByte('\n') // add line break
			out.WriteString(word)
			lineLen = wordlen
		}
	}

	return out.String()
}

// Truncate truncates using rune length from user input
func Truncate(input string, limit int) string {
	r := []rune(input)
	if len(r) <= limit {
		return input
	}
	if limit > 3 {
		return string(r[:limit-3]) + "..."
	}
	return string(r[:limit])
}
