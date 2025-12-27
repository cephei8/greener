# Query Language

## Basics
Query has optional parts: `[matching] [grouping] [group selector] [modifiers]`.
- **Matching part**: Filters testcases based on field values, labels, and status
- **Grouping part**: Groups matching results by session or labels
- **Group selector**: Selects a specific group from grouped results
- **Modifiers**: Pagination (offset/limit) and date range filtering

Examples:

- `status = "pass"`
- `status = "fail" AND #"feature-x" = "on"`
- `#"ci"` (matches testcases with label "ci")
- `!#"flaky"` (matches testcases without label "flaky")
- `status = "skip" group_by(session_id)`
- `group_by(#"os", #"version")`
- `group_by(#"os", #"version") group = ("linux", "2.0.0")`
- `status = "pass" offset = 10 limit = 50`
- `start_date = "2025/01/01 00:00:00" end_date = "2025/12/31 23:59:59"`

## Supported identifiers
| Identifier  | Description          |
|:------------|:---------------------|
| id          | Testcase ID (UUID)   |
| name        | Testcase name        |
| session_id  | Session ID (UUID)    |
| status      | Testcase status      |
| classname   | Test class name      |
| testsuite   | Test suite name      |
| file        | Test file path       |
| #"<label\>" | Label (with value)   |
| #"<label\>" | Label (presence)     |
| !#"<label\>"| Label (absence)      |

## Status values
Valid status values: `"pass"`, `"fail"`, `"error"`, `"skip"`

## Modifiers
| Modifier    | Format                        | Description           |
|:------------|:------------------------------|:----------------------|
| offset      | `offset = <number>`           | Skip N results        |
| limit       | `limit = <number>`            | Return max N results  |
| start_date  | `start_date = "YYYY/MM/DD HH:MM:SS"` | Filter from date |
| end_date    | `end_date = "YYYY/MM/DD HH:MM:SS"`   | Filter to date   |

## Grammar
``` ebnf
query            = base_query modifier_list

base_query       = matching_part?

matching_part    = condition ( logical_op condition )*

logical_op       = "AND" | "OR"

condition        = field_condition
                 | tag_condition
                 | tag_presence
                 | tag_absence

field_condition  = field_ident equality_op quoted_string

field_ident      = "id"
                 | "name"
                 | "session_id"
                 | "status"
                 | "classname"
                 | "testsuite"
                 | "file"

tag_condition    = "#" quoted_label equality_op quoted_string

tag_presence     = "#" quoted_label

tag_absence      = "!" "#" quoted_label

equality_op      = "=" | "!="

quoted_label     = "\"" non_empty_string "\""
quoted_string    = "\"" non_empty_string "\""

modifier_list    = modifier*

modifier         = grouping_part
                 | group_selector
                 | offset_clause
                 | limit_clause
                 | start_date_clause
                 | end_date_clause

grouping_part    = "group_by(" grouping_ident_list ")"

grouping_ident_list = grouping_ident ( "," grouping_ident )*

grouping_ident   = "session_id"
                 | "#" quoted_label

group_selector   = "group" "=" "(" string_list ")"

string_list      = quoted_string ( "," quoted_string )*

offset_clause    = "offset" "=" number

limit_clause     = "limit" "=" number

start_date_clause = "start_date" "=" quoted_datetime

end_date_clause  = "end_date" "=" quoted_datetime

quoted_datetime  = "\"" "YYYY/MM/DD HH:MM:SS" "\""
```
