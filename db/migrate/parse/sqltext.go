package parse

import (
	"strings"
)

// skipQuoted returns the index just past the quoted literal or identifier
// that starts at s[i], treating a doubled quote as an escaped one. An
// unterminated quote runs to the end of s.
func skipQuoted(s string, i int) int {
	quote := s[i]
	for j := i + 1; j < len(s); j++ {
		if s[j] != quote {
			continue
		}
		if j+1 < len(s) && s[j+1] == quote {
			j++
			continue
		}
		return j + 1
	}
	return len(s)
}

func isQuote(c byte) bool {
	return c == '\'' || c == '"'
}

// maskQuoted returns s with the contents of quoted literals and identifiers
// replaced by spaces, keeping every byte offset, so keyword checks do not
// match text inside a default such as 'NOT NULL'.
func maskQuoted(s string) string {
	masked := []byte(s)
	for i := 0; i < len(s); {
		if !isQuote(s[i]) {
			i++
			continue
		}
		end := skipQuoted(s, i)
		for j := i + 1; j < end-1; j++ {
			masked[j] = ' '
		}
		i = end
	}
	return string(masked)
}

func isIdentifierByte(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

// defaultExpression returns the expression after the DEFAULT keyword of a
// column definition, or false when it has none.
func defaultExpression(s string) (string, bool) {
	masked := maskQuoted(s)
	upper := strings.ToUpper(masked)
	for from := 0; ; {
		idx := strings.Index(upper[from:], "DEFAULT")
		if idx < 0 {
			return "", false
		}
		start := from + idx
		end := start + len("DEFAULT")
		from = end
		if start > 0 && isIdentifierByte(masked[start-1]) {
			continue
		}
		if end >= len(s) || !isSpace(s[end]) {
			continue
		}
		return readExpression(strings.TrimLeft(s[end:], " \t\r\n\f\v")), true
	}
}

// readExpression returns the SQL expression at the start of s: a quoted
// literal, a parenthesized expression or a token such as now() or 1.5, with
// any suffix up to the next space, comma or semicolon outside quotes and
// parentheses.
func readExpression(s string) string {
	depth := 0
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case isQuote(c):
			i = skipQuoted(s, i)
			continue
		case c == '(':
			depth++
		case c == ')':
			if depth == 0 {
				return s[:i]
			}
			depth--
		case depth == 0 && (c == ',' || c == ';' || isSpace(c)):
			return s[:i]
		}
		i++
	}
	return s
}

// unquoteLiteral returns the value of a single-quoted SQL string literal, or
// false when s is not exactly one.
func unquoteLiteral(s string) (string, bool) {
	if len(s) < 2 || s[0] != '\'' || skipQuoted(s, 0) != len(s) || s[len(s)-1] != '\'' {
		return "", false
	}
	return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), true
}
