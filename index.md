---
title: Greener Documentation
---

Greener is a lean and mean test result explorer.

Among other use cases, it lets you:

- Get to test results fast (query specific sessions, tests, statuses, labels etc.)
- Group test results and check aggregated statuses (e.g. `group_by(#"os", #"version")` labels)

Features:

- Easy to use
- No changes to test code needed
- Simple SQL-like query language (with grouping support)
- MCP server for AI agent integration
- Attach labels and/or baggage (arbitrary JSON) to test sessions
- Self-contained executable (only requires SQLite/PostgreSQL/MySQL database)
- Small (~27mb executable / compressed Docker image)

Demo:
![Demo](docs/content/assets/demo.gif)

Demo of AI agent using Greener via MCP:
![Demo - MCP](docs/content/assets/demo_mcp.gif)

# Get Started

## Platform setup
Pull and run the Docker image [cephei8/greener](https://hub.docker.com/r/cephei8/greener).

The following environment variables must be specified:

- `GREENER_DATABASE_URL`: database url, e.g. `postgres://postgres:qwerty@localhost:5432/postgres?ssl=disable`
- `GREENER_AUTH_SECRET` - JWT secret, e.g. `abcdefg1234567`

### Example
Mount local directory and run Greener with SQLite database:
``` shell
mkdir greener-data

docker run --rm \
    -v $(pwd)/greener-data:/app/data \
    -p 8080:8080 \
    -e GREENER_DATABASE_URL=sqlite:////app/data/greener.db \
    -e GREENER_AUTH_SECRET=my-secret \
    cephei8/greener:latest
```

Create user:
``` shell
go install git.sr.ht/~cephei8/greener/server/cmd/greener-admin@main

greener-admin \
    --db-url sqlite:///greener-data/greener.db \
    create-user \
    --username greener \
    --password greener \
    --role editor
```

The `--role` flag accepts `editor` (full access) or `viewer` (read-only). Default is `viewer`.

Now, access Greener at http://localhost:8080 in your browser.

## Reporting test results to Greener
Check out [Ecosystem](ecosystem.md) for ways to report test results to Greener.

# Platform Configuration
| Environment Variable                    | Is Required? | Description                                         | Value Example                             |
|:----------------------------------------|:-------------|:----------------------------------------------------|:------------------------------------------|
| GREENER_DATABASE_URL                    | *Yes*        | Database URL                                        | `postgres://postgres:qwerty@db:5432/postgres` |
| GREENER_PORT                            | No           | Port to listen on (default: 8080)                   | `8080`                                    |
| GREENER_AUTH_SECRET                     | *Yes*        | JWT secret                                          | `abcdefg1234567`                          |
| GREENER_AUTH_ISSUER                     | No           | External base URL (for OAuth, defaults to localhost)| `https://greener.example.com`             |
| GREENER_ALLOW_UNAUTHENTICATED_VIEWERS   | No           | Allow unauthenticated users to view data (read-only)| `true`                                    |

## User Roles

Greener supports two user roles:

- **editor**: Full access to all features including creating API keys
- **viewer**: Read-only access to test results

When creating users with `greener-admin`, specify the role with `--role`:

```shell
greener-admin --db-url <url> create-user --username <user> --password <pass> --role editor
```

The default role is `viewer`.

# Plugin Configuration

| Environment Variable        | Is Required? | Description                     | Value Example                            |
|:----------------------------|:-------------|:--------------------------------|:-----------------------------------------|
| GREENER_INGRESS_ENDPOINT    | *Yes*        | Server URL                      | `http://localhost:5096`                  |
| GREENER_INGRESS_API_KEY     | *Yes*        | API key                         | \[API key created in Greener\]           |
| GREENER_SESSION_ID          | *No*         | Session UUIDv4 ID               | `"b7e499fd-f6e1-435c-8ef7-624287ca2bd4"` |
| GREENER_SESSION_DESCRIPTION | *No*         | Session description             | `"My test session"`                      |
| GREENER_SESSION_LABELS      | *No*         | Labels to attach to the session | `"label1=value1,label2"`                 |
| GREENER_SESSION_BAGGAGE     | *No*         | JSON to attach to the session   | `'{"version": "2.0.0"}'`                 |

# Ecosystem

## Test framework plugins
| Programming Language | Framework | Package                                                      | Repository                                                                    |
|:---------------------|:----------|:-------------------------------------------------------------|:------------------------------------------------------------------------------|
| Python               | pytest    | [pytest-greener](https://pypi.org/project/pytest-greener/)   | [~cephei8/pytest-greener](https://git.sr.ht/~cephei8/pytest-greener)           |
| JavaScript           | Jest      | [jest-greener](https://www.npmjs.com/package/jest-greener)   | [~cephei8/jest-greener](https://git.sr.ht/~cephei8/jest-greener)               |
| JavaScript           | Mocha     | [mocha-greener](https://www.npmjs.com/package/mocha-greener) | [~cephei8/mocha-greener](https://git.sr.ht/~cephei8/mocha-greener)             |
| Go                   | N/A       | N/A                                                          | [~cephei8/greener-reporter-go](https://git.sr.ht/~cephei8/greener-reporter-go) |

## Generic
- Tool to report JUnit XML results: [~cephei8/greener-reporter-junitxml](https://git.sr.ht/~cephei8/greener-reporter-junitxml)
- CLI tool to report test results: [~cephei8/greener-reporter-cli](https://git.sr.ht/~cephei8/greener-reporter-cli)

## Supporting libraries
- Python library for implementing reporters: [~cephei8/greener-reporter-py](https://git.sr.ht/~cephei8/greener-reporter-py)
- JavaScript library for implementing reporters: [~cephei8/greener-reporter-js](https://git.sr.ht/~cephei8/greener-reporter-js)
- C FFI library for implementing reporters: [~cephei8/greener-reporter](https://git.sr.ht/~cephei8/greener-reporter)

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

# MCP Server

Greener includes an MCP (Model Context Protocol) server that allows AI agents to query test results.

The MCP server is available at `/api/v1/mcp` and uses OAuth 2.0 for authentication.

## How It Works

When an AI agent queries test results through the MCP server, you can view the results in two ways:

1. **In the agent interface** - The agent receives the query results directly and can analyze, summarize, or discuss them with you
2. **In the Greener UI** - Queries can trigger a server-sent event (SSE) that updates the Greener browser tab to display the same results

This means you can ask your AI agent questions like "show me all failing tests from today" and see the results both in your conversation and in the Greener web interface simultaneously.

To disable the browser update, agents can set `trigger_sse: false` when calling MCP tools. Just tell the agent how you prefer to see the results.

## Available Tools

| Tool              | Description                                           |
|:------------------|:------------------------------------------------------|
| query_testcases   | Query test cases using the Greener query language     |
| query_sessions    | Query test sessions using the Greener query language  |
| query_groups      | Query grouped test results with group_by clause       |
| get_testcase      | Get detailed information about a specific test case   |
| get_session       | Get detailed information about a specific test session|

## OAuth Endpoints

| Endpoint                                  | Description                       |
|:------------------------------------------|:----------------------------------|
| `/.well-known/oauth-authorization-server` | OAuth server metadata             |
| `/oauth/authorize`                        | Authorization endpoint            |
| `/oauth/token`                            | Token endpoint                    |
| `/oauth/register`                         | Dynamic client registration       |

## Configuration

Set `GREENER_AUTH_ISSUER` environment variable to your external base URL if it differs from localhost.
