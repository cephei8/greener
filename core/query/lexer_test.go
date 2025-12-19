package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer_EmptyInput(t *testing.T) {
	lexer := newQueryLexer("")
	var lval yySymType
	token := lexer.Lex(&lval)
	assert.Equal(t, 0, token, "expected EOF token")
}

func TestLexer_Operators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"equals", "=", EQUALS},
		{"not equals", "!=", NOTEQUALS},
		{"hash", "#", HASH},
		{"bang", "!", BANG},
		{"comma", ",", COMMA},
		{"left paren", "(", LPAREN},
		{"right paren", ")", RPAREN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := newQueryLexer(tt.input)
			var lval yySymType
			token := lexer.Lex(&lval)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"and", "and", AND},
		{"AND uppercase", "AND", AND},
		{"or", "or", OR},
		{"OR uppercase", "OR", OR},
		{"session_id", "session_id", SESSION_ID},
		{"SESSION_ID uppercase", "SESSION_ID", SESSION_ID},
		{"id", "id", ID},
		{"name", "name", NAME},
		{"classname", "classname", CLASSNAME},
		{"testsuite", "testsuite", TESTSUITE},
		{"file", "file", FILE},
		{"status", "status", STATUS},
		{"group_by", "group_by", GROUP_BY},
		{"group", "group", GROUP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := newQueryLexer(tt.input)
			var lval yySymType
			token := lexer.Lex(&lval)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestLexer_Strings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", `"hello"`, "hello"},
		{"empty string", `""`, ""},
		{"string with spaces", `"hello world"`, "hello world"},
		{"string with special chars", `"test-123_abc"`, "test-123_abc"},
		{"unterminated string", `"hello`, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := newQueryLexer(tt.input)
			var lval yySymType
			token := lexer.Lex(&lval)
			assert.Equal(t, STRING, token)
			assert.Equal(t, tt.expected, lval.String)
		})
	}
}

func TestLexer_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"leading whitespace", "  =", EQUALS},
		{"trailing whitespace", "=  ", EQUALS},
		{"tabs", "\t\t=", EQUALS},
		{"newlines", "\n\n=", EQUALS},
		{"mixed whitespace", " \t\n\r =", EQUALS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := newQueryLexer(tt.input)
			var lval yySymType
			token := lexer.Lex(&lval)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestLexer_MultipleTokens(t *testing.T) {
	input := `name = "test" and status = "passed"`
	lexer := newQueryLexer(input)
	var lval yySymType

	expectedTokens := []struct {
		token int
		value string
	}{
		{NAME, ""},
		{EQUALS, ""},
		{STRING, "test"},
		{AND, ""},
		{STATUS, ""},
		{EQUALS, ""},
		{STRING, "passed"},
		{0, ""}, // EOF
	}

	for i, expected := range expectedTokens {
		token := lexer.Lex(&lval)
		assert.Equal(t, expected.token, token, "token %d", i)
		if expected.value != "" {
			assert.Equal(t, expected.value, lval.String, "token %d value", i)
		}
	}
}

func TestLexer_GroupByQuery(t *testing.T) {
	input := `group_by(session_id, status)`
	lexer := newQueryLexer(input)
	var lval yySymType

	expectedTokens := []int{GROUP_BY, LPAREN, SESSION_ID, COMMA, STATUS, RPAREN, 0}

	for i, expected := range expectedTokens {
		token := lexer.Lex(&lval)
		assert.Equal(t, expected, token, "token %d", i)
	}
}

func TestLexer_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unknown identifier", "unknown_keyword"},
		{"unexpected character", "@"},
		{"invalid character", "$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := newQueryLexer(tt.input)
			var lval yySymType
			token := lexer.Lex(&lval)
			assert.Equal(t, 0, token, "expected error token")
			assert.NotNil(t, lexer.err, "expected error to be set")
		})
	}
}

func TestLexer_ComplexQuery(t *testing.T) {
	input := `(name = "test1" or name = "test2") and status != "failed"`
	lexer := newQueryLexer(input)
	var lval yySymType

	expectedTokens := []struct {
		token int
		value string
	}{
		{LPAREN, ""},
		{NAME, ""},
		{EQUALS, ""},
		{STRING, "test1"},
		{OR, ""},
		{NAME, ""},
		{EQUALS, ""},
		{STRING, "test2"},
		{RPAREN, ""},
		{AND, ""},
		{STATUS, ""},
		{NOTEQUALS, ""},
		{STRING, "failed"},
		{0, ""},
	}

	for i, expected := range expectedTokens {
		token := lexer.Lex(&lval)
		assert.Equal(t, expected.token, token, "token %d", i)
		if expected.value != "" {
			assert.Equal(t, expected.value, lval.String, "token %d value", i)
		}
	}
}
