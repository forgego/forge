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
	start, end, ok := defaultSpan(s)
	if !ok {
		return "", false
	}
	return s[start:end], true
}

// maskedClauses returns the clauses of a column definition after its type
// with quoted text and the default expression blanked out, so keyword checks
// match neither a literal such as 'NOT NULL' nor an expression default, and
// the default expression, if any.
func maskedClauses(s string) (masked string, defaultExpr string, hasDefault bool) {
	masked = maskQuoted(s)
	start, end, ok := defaultSpan(s)
	if !ok {
		return masked, "", false
	}
	return masked[:start] + strings.Repeat(" ", end-start) + masked[end:], s[start:end], true
}

// defaultSpan returns the byte range of the expression after the DEFAULT
// keyword of a column definition, or false when it has none.
func defaultSpan(s string) (int, int, bool) {
	masked := maskQuoted(s)
	upper := strings.ToUpper(masked)
	for from := 0; ; {
		idx := strings.Index(upper[from:], "DEFAULT")
		if idx < 0 {
			return 0, 0, false
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
		for end < len(s) && isSpace(s[end]) {
			end++
		}
		return end, end + len(readExpression(s[end:])), true
	}
}

// columnConstraintKeywords end a column default: they start the next clause
// of a column definition. NOT ends it only when NULL follows.
var columnConstraintKeywords = map[string]bool{
	"NULL": true, "UNIQUE": true, "PRIMARY": true, "CHECK": true, "REFERENCES": true,
	"GENERATED": true, "COLLATE": true, "CONSTRAINT": true,
}

// readExpression returns the SQL expression at the start of s, such as a
// quoted literal, a function call, a cast ('{}'::jsonb) or an operator
// expression ('x' || 'y'). It ends before a comma, a semicolon or an
// unbalanced closing parenthesis, or before a column-constraint keyword (NOT
// NULL, NULL, UNIQUE, PRIMARY, CHECK, REFERENCES, GENERATED, COLLATE,
// CONSTRAINT), outside quotes and parentheses. Trailing spaces are not part of
// it.
func readExpression(s string) string {
	depth := 0
	end := 0
	// caseDepth counts open CASE ... END blocks, inside which NULL and NOT are
	// operands (THEN NULL), not column constraints. prev and prev2 are the
	// previous two words, so IS NULL and IS NOT NULL continue the expression.
	caseDepth := 0
	prev, prev2 := "", ""
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case isQuote(c):
			i = skipQuoted(s, i)
			end = i
			continue
		case c == '(':
			depth++
		case c == ')':
			if depth == 0 {
				return s[:end]
			}
			depth--
		case depth == 0 && (c == ',' || c == ';'):
			return s[:end]
		case depth == 0 && isIdentifierByte(c) && (i == 0 || !isIdentifierByte(s[i-1])):
			j := i
			for j < len(s) && isIdentifierByte(s[j]) {
				j++
			}
			word := strings.ToUpper(s[i:j])
			// The expression's first word is never a keyword that ends it,
			// so DEFAULT NULL reads as NULL.
			if end > 0 && caseDepth == 0 && !continuesOperand(word, prev, prev2, s[:i]) &&
				endsExpression(word, s[j:]) {
				return s[:end]
			}
			switch {
			case word == "CASE":
				caseDepth++
			case word == "END" && caseDepth > 0:
				caseDepth--
			}
			prev2, prev = prev, word
			i = j
			end = i
			continue
		}
		i++
		if !isSpace(c) {
			end = i
		}
	}
	return s[:end]
}

// continuesOperand reports whether NULL or NOT at this point is part of the
// expression: after IS (x IS NULL, x IS NOT NULL), after IS NOT, or right
// after an operator such as = or ||.
func continuesOperand(word, prev, prev2, before string) bool {
	if word != "NULL" && word != "NOT" {
		return false
	}
	if prev == "IS" || (word == "NULL" && prev == "NOT" && prev2 == "IS") {
		return true
	}
	trimmed := strings.TrimRight(before, " \t\r\n\f\v")
	if trimmed == "" {
		return false
	}
	return strings.ContainsRune("=<>!+-*/%|&^~(,", rune(trimmed[len(trimmed)-1]))
}

// endsExpression reports whether word, followed by rest, starts a column
// constraint rather than continuing an expression.
func endsExpression(word, rest string) bool {
	if word == "NOT" {
		next := strings.TrimLeft(rest, " \t\r\n\f\v")
		return len(next) >= 4 && strings.EqualFold(next[:4], "NULL") && (len(next) == 4 || !isIdentifierByte(next[4]))
	}
	return columnConstraintKeywords[word]
}

// unquoteLiteral returns the value of a single-quoted SQL string literal, or
// false when s is not exactly one.
func unquoteLiteral(s string) (string, bool) {
	if len(s) < 2 || s[0] != '\'' || skipQuoted(s, 0) != len(s) || s[len(s)-1] != '\'' {
		return "", false
	}
	return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), true
}

// stripComments replaces -- and /* */ comments outside quoted text with a
// space, keeping the statement's structure.
func stripComments(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		switch {
		case isQuote(s[i]):
			end := skipQuoted(s, i)
			b.WriteString(s[i:end])
			i = end
		case strings.HasPrefix(s[i:], "--"):
			end := strings.IndexByte(s[i:], '\n')
			if end < 0 {
				return b.String()
			}
			b.WriteByte(' ')
			i += end
		case strings.HasPrefix(s[i:], "/*"):
			end := strings.Index(s[i+2:], "*/")
			if end < 0 {
				return b.String()
			}
			b.WriteByte(' ')
			i += end + 4
		default:
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
