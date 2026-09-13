package parse

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Statement represents a scanned SQL statement with its position and comments
type Statement struct {
	Pos      int      // statement position in file
	Text     string   // statement text (trimmed)
	Comments []string // associated comments
}

// Lexer scans SQL and splits it into individual statements
type Lexer struct {
	// Scanner state
	src, input string   // src and current input text
	pos        int      // current position in input
	total      int      // total bytes scanned so far
	width      int      // size of latest rune
	delim      string   // statement delimiter (default ";")
	comments   []string // collected comments

	// Options
	matchDollarQuote bool // PostgreSQL dollar-quoted strings
	matchBegin       bool // BEGIN/END blocks
	hashComments     bool // MySQL hash comments (#)
}

// LexerOptions controls lexer behavior
type LexerOptions struct {
	MatchDollarQuote bool // Enable PostgreSQL dollar-quoted strings
	MatchBegin       bool // Enable BEGIN/END block handling
	HashComments     bool // Enable MySQL hash comments
}

// NewLexer creates a new SQL statement lexer with default options
func NewLexer() *Lexer {
	return NewLexerWithOptions(LexerOptions{
		MatchDollarQuote: true, // Enable by default for PostgreSQL
		MatchBegin:       true, // Enable by default
		HashComments:     false,
	})
}

// NewLexerWithOptions creates a new lexer with custom options
func NewLexerWithOptions(opts LexerOptions) *Lexer {
	return &Lexer{
		delim:            ";",
		matchDollarQuote: opts.MatchDollarQuote,
		matchBegin:       opts.MatchBegin,
		hashComments:     opts.HashComments,
	}
}

// Scan scans the input SQL and returns all statements
func (l *Lexer) Scan(input string) ([]*Statement, error) {
	var stmts []*Statement
	if err := l.init(input); err != nil {
		return nil, err
	}

	for {
		stmt, err := l.nextStatement()
		if err == io.EOF {
			return stmts, nil
		}
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
}

// init initializes the lexer state
func (l *Lexer) init(input string) error {
	l.comments = nil
	l.pos, l.total, l.width = 0, 0, 0
	l.src, l.input = input, input
	l.delim = ";"
	return nil
}

const eos = -1 // end of string

var (
	// Dollar-quoted string pattern: $tag$...$tag$
	reDollarQuote = regexp.MustCompile(`^\$([A-Za-z_][\w]*)*\$`)
	// BEGIN block pattern
	reBegin = regexp.MustCompile(`(?i)^\s*BEGIN\s+`)
	reEnd   = regexp.MustCompile(`(?i)^\s*END\s*`)
)

// nextStatement scans the next SQL statement
func (l *Lexer) nextStatement() (*Statement, error) {
	var (
		depth, openingPos int
		text              string
	)

	l.skipSpaces()

Scan:
	for {
		switch r := l.next(); {
		case r == eos:
			switch {
			case depth > 0:
				return nil, l.error(openingPos, "unclosed '('")
			case l.pos > 0:
				text = l.input
				break Scan
			default:
				return nil, io.EOF
			}
		case r == '(':
			if depth == 0 {
				openingPos = l.pos
			}
			depth++
		case r == ')':
			if depth == 0 {
				return nil, l.error(l.pos, "unexpected ')'")
			}
			depth--
		case r == '\'', r == '"', r == '`':
			if err := l.skipQuote(r); err != nil {
				return nil, err
			}
		// Delimiters take precedence over comments
		case depth == 0 && l.pos >= l.width && strings.HasPrefix(l.input[l.pos-l.width:], l.delim):
			l.addPos(len(l.delim) - l.width)
			text = l.input[:l.pos]
			break Scan
		case l.matchDollarQuote && r == '$' && l.pos > 0 && reDollarQuote.MatchString(l.input[l.pos-1:]):
			if err := l.skipDollarQuote(); err != nil {
				return nil, err
			}
		case r == '#' && l.hashComments:
			l.comment("#", "\n")
		case r == '-' && l.peek() == '-':
			l.next() // consume second '-'
			l.comment("--", "\n")
		case r == '/' && l.peek() == '*':
			l.next() // consume '*'
			l.comment("/*", "*/")
		case l.matchBegin && l.delim == ";" && l.pos > 0 && reBegin.MatchString(l.input[l.pos-1:]):
			if err := l.skipBegin(); err == nil {
				text = l.input[:l.pos]
				break Scan
			}
			// Not a BEGIN block, continue
		}
	}

	return l.emit(text), nil
}

// next returns the next rune and advances position
func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		return eos
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.addPos(w)
	return r
}

// peek returns the next rune without advancing position
func (l *Lexer) peek() rune {
	p, w, t := l.pos, l.width, l.total
	r := l.next()
	l.pos, l.width, l.total = p, w, t
	return r
}

// addPos advances position
func (l *Lexer) addPos(p int) {
	l.pos += p
	l.total += p
}

// skipQuote skips a quoted string (single, double, or backtick)
func (l *Lexer) skipQuote(quote rune) error {
	pos := l.pos
	for {
		switch r := l.next(); {
		case r == eos:
			return l.error(pos, "unclosed quote %q", quote)
		case r == '\\':
			// Skip escaped character
			l.next()
		case r == quote:
			// Check for escaped quote (double quote)
			if quote == '\'' && l.peek() == '\'' {
				l.next() // consume second quote
				continue
			}
			return nil
		}
	}
}

// skipDollarQuote skips a PostgreSQL dollar-quoted string
func (l *Lexer) skipDollarQuote() error {
	if l.pos == 0 {
		return l.error(l.pos, "unexpected dollar quote")
	}
	m := reDollarQuote.FindString(l.input[l.pos-1:])
	if m == "" {
		return l.error(l.pos, "unexpected dollar quote")
	}
	l.addPos(len(m) - 1) // already consumed '$'

	// Find matching closing tag
	for {
		switch r := l.next(); {
		case r == eos:
			return l.error(l.pos, "unclosed dollar-quoted string")
		case r == '$' && strings.HasPrefix(l.input[l.pos-1:], m):
			l.addPos(len(m) - 1)
			return nil
		}
	}
}

// skipBegin skips a BEGIN/END block (treats as atomic statement)
func (l *Lexer) skipBegin() error {
	if l.pos == 0 {
		return l.error(l.pos, "not a BEGIN block")
	}
	m := reBegin.FindString(l.input[l.pos-1:])
	if m == "" {
		return l.error(l.pos, "not a BEGIN block")
	}
	l.addPos(len(m) - 1) // already consumed 'B'

	// Scan until matching END
	group := &Lexer{
		matchDollarQuote: l.matchDollarQuote,
		matchBegin:       false, // Don't recurse into nested BEGIN
		hashComments:     l.hashComments,
	}
	if err := group.init(l.input[l.pos:]); err != nil {
		return err
	}

	for {
		stmt, err := group.nextStatement()
		if err == io.EOF {
			return l.error(l.pos, "unexpected EOF in BEGIN block")
		}
		if err != nil {
			return l.error(l.pos, "error scanning BEGIN block: %v", err)
		}

		// Check if this statement is END
		stmtText := strings.TrimSpace(stmt.Text)
		if reEnd.MatchString(stmtText) {
			// Check if END is followed by delimiter or is standalone
			remaining := l.input[l.pos+group.total:]
			remaining = strings.TrimSpace(remaining)
			if strings.HasPrefix(remaining, l.delim) || len(remaining) == 0 {
				l.addPos(group.total)
				// Consume delimiter if present
				if strings.HasPrefix(remaining, l.delim) {
					l.addPos(len(l.delim))
				}
				return nil
			}
		}
		l.addPos(group.total)
	}
}

// comment handles comments (line or block)
func (l *Lexer) comment(left, right string) {
	i := strings.Index(l.input[l.pos:], right)
	if i == -1 {
		// Comment not closed, treat rest as comment
		// This is a warning condition - unclosed comments can cause parsing issues
		// Note: Warning would be logged if verbose mode is enabled in the parser
		l.addPos(len(l.input) - l.pos)
		return
	}

	// If comment is at start of statement, collect it
	if l.pos == len(left) {
		l.addPos(i + len(right))
		l.comments = append(l.comments, l.input[:l.pos])
		l.input = l.input[l.pos:]
		l.pos = 0
		// Double newline separates comment group from statement
		if strings.HasPrefix(l.input, "\n\n") || (right == "\n" && strings.HasPrefix(l.input, "\n")) {
			l.comments = nil
		}
		l.skipSpaces()
	} else {
		// Comment inside statement, just skip it
		l.addPos(i + len(right))
	}
}

// skipSpaces skips whitespace
func (l *Lexer) skipSpaces() {
	n := len(l.input)
	l.input = strings.TrimLeftFunc(l.input, unicode.IsSpace)
	l.total += n - len(l.input)
}

// emit creates a Statement from the scanned text
func (l *Lexer) emit(text string) *Statement {
	pos := l.total - len(text)
	if pos < 0 {
		pos = 0
	}
	stmt := &Statement{
		Pos:      pos,
		Text:     strings.TrimSpace(text),
		Comments: l.comments,
	}

	// Remove delimiter from text if present
	if strings.HasSuffix(stmt.Text, l.delim) {
		stmt.Text = strings.TrimSuffix(stmt.Text, l.delim)
		stmt.Text = strings.TrimSpace(stmt.Text)
	}

	// Advance input past this statement
	l.input = l.input[l.pos:]
	l.pos = 0
	l.comments = nil

	return stmt
}

// error creates an error with position information
func (l *Lexer) error(pos int, format string, args ...interface{}) error {
	format = "%d:%d: " + format
	var (
		p    = len(l.src) - len(l.input) + pos
		line = 1
		col  = p
	)
	// Ensure p is within bounds
	if p > len(l.src) {
		p = len(l.src)
	}
	if p < 0 {
		p = 0
	}
	src := l.src[:p]
	lastNewline := strings.LastIndex(src, "\n")
	if lastNewline >= 0 {
		line = 1 + strings.Count(src, "\n")
		col = p - lastNewline - 1
	}
	return fmt.Errorf(format, append([]interface{}{line, col}, args...)...)
}
