%{
package query

import (
	"fmt"
	"github.com/google/uuid"
)

%}

%union {
	SelectQuery         SelectQuery
	CompoundSelectQuery CompoundSelectQuery
	LogicalOperator     LogicalOperator
	EqualityOperator    EqualityOperator
	String              string
	Strings             []string
	GroupQuery          GroupQuery
	GroupSelector       []string
	GroupToken          GroupToken
	GroupTokens         []GroupToken
	Number              int
}

%token <String> STRING IDENTIFIER
%token <Number> NUMBER
%token EQUALS NOTEQUALS
%token AND OR
%token HASH BANG COMMA LPAREN RPAREN
%token SESSION_ID ID NAME CLASSNAME TESTSUITE FILE STATUS GROUP_BY GROUP OFFSET LIMIT

%type <SelectQuery> query atomic_query field_query tag_query not_tag_query
%type <CompoundSelectQuery> compound_query
%type <LogicalOperator> logical_op
%type <EqualityOperator> equality_op
%type <GroupQuery> group_query
%type <GroupSelector> group_selector
%type <GroupToken> group_token
%type <GroupTokens> group_token_list
%type <Strings> string_list

%%

query:
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
        }
	}
	| OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			Offset: $3,
        }
	}
	| LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			Limit: $3,
        }
	}
	| OFFSET EQUALS NUMBER LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			Offset: $3,
			Limit:  $6,
        }
	}
	| LIMIT EQUALS NUMBER OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			Offset: $6,
			Limit:  $3,
        }
	}
	| compound_query
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: $1,
        }
	}
	| compound_query OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: $1,
			Offset:      $4,
        }
	}
	| compound_query LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: $1,
			Limit:       $4,
        }
	}
	| compound_query OFFSET EQUALS NUMBER LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: $1,
			Offset:      $4,
			Limit:       $7,
        }
	}
	| compound_query LIMIT EQUALS NUMBER OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: $1,
			Offset:      $7,
			Limit:       $4,
        }
	}
	| group_query
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery:  &$1,
		}
	}
	| group_query OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery: &$1,
			Offset:     $4,
		}
	}
	| group_query LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery: &$1,
			Limit:      $4,
		}
	}
	| group_query OFFSET EQUALS NUMBER LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery: &$1,
			Offset:     $4,
			Limit:      $7,
		}
	}
	| group_query LIMIT EQUALS NUMBER OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery: &$1,
			Offset:     $7,
			Limit:      $4,
		}
	}
	| group_query group_selector
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery:    &$1,
			GroupSelector: $2,
		}
	}
	| group_query group_selector OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery:    &$1,
			GroupSelector: $2,
			Offset:        $5,
		}
	}
	| group_query group_selector LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery:    &$1,
			GroupSelector: $2,
			Limit:         $5,
		}
	}
	| group_query group_selector OFFSET EQUALS NUMBER LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery:    &$1,
			GroupSelector: $2,
			Offset:        $5,
			Limit:         $8,
		}
	}
	| group_query group_selector LIMIT EQUALS NUMBER OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
            SelectQuery: CompoundSelectQuery{ Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart { Operator: OpAnd, Query: EmptySelectQuery{} },
            }},
			GroupQuery:    &$1,
			GroupSelector: $2,
			Offset:        $8,
			Limit:         $5,
		}
	}
	| compound_query group_query
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery: $1,
			GroupQuery:  &$2,
		}
	}
	| compound_query group_query OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery: $1,
			GroupQuery:  &$2,
			Offset:      $5,
		}
	}
	| compound_query group_query LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery: $1,
			GroupQuery:  &$2,
			Limit:       $5,
		}
	}
	| compound_query group_query OFFSET EQUALS NUMBER LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery: $1,
			GroupQuery:  &$2,
			Offset:      $5,
			Limit:       $8,
		}
	}
	| compound_query group_query LIMIT EQUALS NUMBER OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery: $1,
			GroupQuery:  &$2,
			Offset:      $8,
			Limit:       $5,
		}
	}
	| compound_query group_query group_selector
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery:   $1,
			GroupQuery:    &$2,
			GroupSelector: $3,
		}
	}
	| compound_query group_query group_selector OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery:   $1,
			GroupQuery:    &$2,
			GroupSelector: $3,
			Offset:        $6,
		}
	}
	| compound_query group_query group_selector LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery:   $1,
			GroupQuery:    &$2,
			GroupSelector: $3,
			Limit:         $6,
		}
	}
	| compound_query group_query group_selector OFFSET EQUALS NUMBER LIMIT EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery:   $1,
			GroupQuery:    &$2,
			GroupSelector: $3,
			Offset:        $6,
			Limit:         $9,
		}
	}
	| compound_query group_query group_selector LIMIT EQUALS NUMBER OFFSET EQUALS NUMBER
	{
		yylex.(*queryLexer).result = Query{
			SelectQuery:   $1,
			GroupQuery:    &$2,
			GroupSelector: $3,
			Offset:        $9,
			Limit:         $6,
		}
	}
	| compound_query group_selector
	{
		yylex.Error("group specification requires group_by clause")
		return 1
	}
	;

compound_query:
	atomic_query
	{
        $$ = CompoundSelectQuery{
            Parts: []CompoundSelectQueryPart{
                CompoundSelectQueryPart{ Operator: OpAnd, Query: $1 },
            },
        }
	}
	| compound_query logical_op atomic_query
	{
        
		$1.Parts = append($1.Parts, CompoundSelectQueryPart{ Operator: $2, Query: $3 })
        $$ = $1
    }
	;

logical_op:
	AND
	{
		$$ = OpAnd
	}
	| OR
	{
		$$ = OpOr
	}
	;

atomic_query:
	field_query
	{
		$$ = $1
	}
	| tag_query
	{
		$$ = $1
	}
	| not_tag_query
	{
		$$ = $1
	}
	;

field_query:
	SESSION_ID equality_op STRING
	{
		sessionId, err := uuid.Parse($3)
		if err != nil {
			yylex.Error(fmt.Sprintf("invalid UUID: %s", $3))
			return 1
		}
		$$ = SessionSelectQuery{
			SessionId: sessionId,
			Operator:  $2,
		}
	}
	| ID equality_op STRING
	{
		id, err := uuid.Parse($3)
		if err != nil {
			yylex.Error(fmt.Sprintf("invalid UUID: %s", $3))
			return 1
		}
		$$ = IdSelectQuery{
			Id:       id,
			Operator: $2,
		}
	}
	| NAME equality_op STRING
	{
		$$ = NameSelectQuery{
			Name:     $3,
			Operator: $2,
		}
	}
	| CLASSNAME equality_op STRING
	{
		$$ = ClassnameSelectQuery{
			Classname: $3,
			Operator:  $2,
		}
	}
	| TESTSUITE equality_op STRING
	{
		$$ = TestsuiteSelectQuery{
			Testsuite: $3,
			Operator:  $2,
		}
	}
	| FILE equality_op STRING
	{
		$$ = FileSelectQuery{
			File:     $3,
			Operator: $2,
		}
	}
	| STATUS equality_op STRING
	{
		validStatuses := []TestcaseStatus{StatusPass, StatusFail, StatusError, StatusSkip}
		var status TestcaseStatus
                isValid := false
		for _, s := range validStatuses {
			if string(s) == $3 {
                                status = s
				isValid = true
				break
			}
		}
		if !isValid {
			yylex.Error(fmt.Sprintf("invalid status: %s", $3))
			return 1
		}
		$$ = StatusSelectQuery{
			Status:   status,
			Operator: $2,
		}
	}
	;

tag_query:
	HASH STRING
	{
		$$ = TagSelectQuery{
			Tag:      $2,
			Operator: OpEq,
		}
	}
	| HASH STRING equality_op STRING
	{
		$$ = TagValueSelectQuery{
			Tag:      $2,
			Value:    $4,
			Operator: $3,
		}
	}
	;

not_tag_query:
	BANG HASH STRING
	{
		$$ = TagSelectQuery{
			Tag:      $3,
			Operator: OpNEq,
		}
	}
	;

equality_op:
	EQUALS
	{
		$$ = OpEq
	}
	| NOTEQUALS
	{
		$$ = OpNEq
	}
	;

group_query:
	GROUP_BY LPAREN group_token_list RPAREN
	{
		$$ = GroupQuery{
			Tokens: $3,
		}
	}
	;

group_token_list:
	group_token
	{
		$$ = []GroupToken{$1}
	}
	| group_token_list COMMA group_token
	{
		$$ = append($1, $3)
	}
	;

group_token:
	SESSION_ID
	{
		$$ = SessionGroupToken{}
	}
	| HASH STRING
	{
		$$ = TagGroupToken{
			Tag: $2,
		}
	}
	;

group_selector:
	GROUP EQUALS LPAREN string_list RPAREN
	{
		$$ = $4
	}
	;

string_list:
	STRING
	{
		$$ = []string{$1}
	}
	| string_list COMMA STRING
	{
		$$ = append($1, $3)
	}
	;

%%
