package query

type Parser struct {
	input string
}

func NewParser(input string) *Parser {
	return &Parser{input: input}
}

func (p *Parser) Parse() (Query, error) {
	lexer := newQueryLexer(p.input)
	result := yyParse(lexer)

	if lexer.err != nil {
		return Query{}, lexer.err
	}
	if result != 0 {
		return Query{}, &QueryError{Message: "unknown parse error"}
	}

	return lexer.result, nil
}
