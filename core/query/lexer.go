package query

import (
	"fmt"
	"strings"
	"unicode"
)

type queryLexer struct {
	input  string
	pos    int
	ch     rune
	result Query
	err    error
}

func newQueryLexer(input string) *queryLexer {
	l := &queryLexer{input: input}
	l.readChar()
	return l
}

func (l *queryLexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
}

func (l *queryLexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *queryLexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *queryLexer) readString() string {
	if l.ch != '"' {
		return ""
	}
	l.readChar()

	var result strings.Builder
	for l.ch != '"' && l.ch != 0 {
		result.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch == '"' {
		l.readChar()
	}

	return result.String()
}

func (l *queryLexer) readIdentifier() string {
	start := l.pos - 1
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[start : l.pos-1]
}

func (l *queryLexer) readNumber() int {
	start := l.pos - 1
	for unicode.IsDigit(l.ch) {
		l.readChar()
	}
	num := 0
	for _, r := range l.input[start : l.pos-1] {
		num = num*10 + int(r-'0')
	}
	return num
}

func (l *queryLexer) Lex(lval *yySymType) int {
	l.skipWhitespace()

	switch l.ch {
	case 0:
		return 0

	case '"':
		lval.String = l.readString()
		return STRING

	case '=':
		l.readChar()
		return EQUALS

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return NOTEQUALS
		}
		l.readChar()
		return BANG

	case '#':
		l.readChar()
		return HASH

	case ',':
		l.readChar()
		return COMMA

	case '(':
		l.readChar()
		return LPAREN

	case ')':
		l.readChar()
		return RPAREN

	default:
		if unicode.IsLetter(l.ch) {
			ident := l.readIdentifier()
			identLower := strings.ToLower(ident)

			switch identLower {
			case "and":
				return AND
			case "or":
				return OR
			case "session_id":
				return SESSION_ID
			case "id":
				return ID
			case "name":
				return NAME
			case "classname":
				return CLASSNAME
			case "testsuite":
				return TESTSUITE
			case "file":
				return FILE
			case "status":
				return STATUS
			case "group_by":
				return GROUP_BY
			case "group":
				return GROUP
			case "offset":
				return OFFSET
			case "limit":
				return LIMIT
			default:
				l.err = fmt.Errorf("unknown identifier: %s", ident)
				return 0
			}
		}

		if unicode.IsDigit(l.ch) {
			lval.Number = l.readNumber()
			return NUMBER
		}

		l.err = fmt.Errorf("unexpected character: %c", l.ch)
		l.readChar()
		return 0
	}
}

func (l *queryLexer) Error(s string) {
	l.err = &QueryError{Message: s}
}
