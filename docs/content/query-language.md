# Query Language

## Basics
Query has two optional parts: `[matching] ... [grouping] ...`.  
Matching part filters based on equality/inequality.  
Grouping part groups matching results.

Examples:

- `status = "pass"`
- `status = "fail" AND #"feature-x" = "on"`
- `status = "skip" group_by(session_id)`
- `group_by(#"build_variant")`

## Supported identifiers
| Identifier  | Description     |
|:------------|:----------------|
| id          | Testcase ID     |
| name        | Testcase name   |
| session_id  | Session ID      |
| status      | Testcase status |
| #"<label\>" | Label           |


## Grammar
``` ebnf
query            = matching_part? grouping_part?

matching_part    = condition ( logical_op condition )*

logical_op       = "and" | "or"

condition        = ident comparator value

ident            = "id"
                 | "name"
                 | "session_id"
                 | "status"
                 | "#" quoted_label

comparator       = "=" | "!="

value            = quoted_string
                 | status_literal

status_literal   = "pass" | "fail" | "skip" | "error"

quoted_label     = "\"" non_empty_string "\""
quoted_string    = "\"" non_empty_string "\""

grouping_part    = "group_by(" grouping_ident_list ")"

grouping_ident_list = grouping_ident ( "," grouping_ident )*

grouping_ident   = "session_id"
                 | "#" quoted_label
```
