# Greener

[![Docker Image Version](https://img.shields.io/docker/v/cephei8/greener)](https://hub.docker.com/r/cephei8/greener)

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
![Demo](./docs/content/assets/demo.gif)

Demo of AI agent using Greener via MCP:
![Demo - MCP](./docs/content/assets/demo_mcp.gif)

## Usage

### Docker

```bash
# run Greener server
mkdir greener-data

docker run --rm \
    -v $(pwd)/greener-data:/app/data \
    -p 8080:8080 \
    -e GREENER_DATABASE_URL=sqlite:////app/data/greener.db \
    -e GREENER_AUTH_SECRET=my-secret \
    cephei8/greener:latest

# create user (--role: editor or viewer, default: viewer)
go install git.sr.ht/~cephei8/greener/server/cmd/greener-admin@main

greener-admin \
    --db-url sqlite:///greener-data/greener.db \
    create-user \
    --username greener \
    --password greener \
    --role editor
```

Then open <http://localhost:8080> in your browser.

#### Docker Compose
Check out [compose.yaml](https://git.sr.ht/~cephei8/greener/tree/main/item/server/compose.yaml) for example of using Docker Compose and PostgreSQL.

### Building from source

```bash
cd server

# build
npm install
npm run build
go build ./cmd/greener

# run Greener server
mkdir greener-data

./greener --db-url "sqlite:///greener-data/greener.db" --auth-secret "my-secret"

# create user (--role: editor or viewer, default: viewer)
go install git.sr.ht/~cephei8/greener/server/cmd/greener-admin@main

greener-admin \
    --db-url sqlite:///greener-data/greener.db \
    create-user \
    --username greener \
    --password greener \
    --role editor
```

Then open <http://localhost:8080> in your browser.

## Platform Configuration

| Environment Variable                    | Is Required? | Description                                         | Value Example                             |
|:----------------------------------------|:-------------|:----------------------------------------------------|:------------------------------------------|
| GREENER_DATABASE_URL                    | *Yes*        | Database URL                                        | `postgres://postgres:qwerty@db:5432/postgres` |
| GREENER_PORT                            | No           | Port to listen on (default: 8080)                   | `8080`                                    |
| GREENER_AUTH_SECRET                     | *Yes*        | JWT secret                                          | `abcdefg1234567`                          |
| GREENER_AUTH_ISSUER                     | No           | External base URL (for OAuth, defaults to localhost)| `https://greener.example.com`             |
| GREENER_ALLOW_UNAUTHENTICATED_VIEWERS   | No           | Allow unauthenticated users to view data (read-only)| `true`                                    |

### User Roles

Greener supports two user roles:

- **editor**: Full access to all features including creating API keys
- **viewer**: Read-only access to test results

When creating users with `greener-admin`, specify the role with `--role`:

```shell
greener-admin --db-url <url> create-user --username <user> --password <pass> --role editor
```

The default role is `viewer`.

## Reporting test results to Greener

Check out [Ecosystem section](#ecosystem) for ways to report test results to Greener.

For the "hello world" the easiest option may be to use [greener-reporter-cli](./reporting/greener-reporter-cli).

## Ecosystem

### Test framework plugins
| Programming Language | Framework | Package                                                      | Repository                                                        |
|:---------------------|:----------|:-------------------------------------------------------------|:------------------------------------------------------------------|
| Python               | pytest    | [pytest-greener](https://pypi.org/project/pytest-greener/)   | [reporting/pytest-greener](./reporting/pytest-greener)             |
| JavaScript           | Jest      | [jest-greener](https://www.npmjs.com/package/jest-greener)   | [reporting/jest-greener](./reporting/jest-greener)                 |
| JavaScript           | Mocha     | [mocha-greener](https://www.npmjs.com/package/mocha-greener) | [reporting/mocha-greener](./reporting/mocha-greener)               |
| Go                   | N/A       | N/A                                                          | [reporting/greener-reporter-go](./reporting/greener-reporter-go)   |

### Generic
- Tool to report JUnit XML results: [greener-reporter-junitxml](./reporting/greener-reporter-junitxml)
- CLI tool to report test results: [greener-reporter-cli](./reporting/greener-reporter-cli)

### Supporting libraries
- Python library for implementing reporters: [greener-reporter-py](./reporting/greener-reporter-py)
- JavaScript library for implementing reporters: [greener-reporter-js](./reporting/greener-reporter-js)
- C FFI library for implementing reporters: [greener-reporter](./reporting/greener-reporter)

## Plugin Configuration

| Environment Variable        | Is Required? | Description                     | Value Example                            |
|:----------------------------|:-------------|:--------------------------------|:-----------------------------------------|
| GREENER_INGRESS_ENDPOINT    | *Yes*        | Server URL                      | `http://localhost:5096`                  |
| GREENER_INGRESS_API_KEY     | *Yes*        | API key                         | \[API key created in Greener\]           |
| GREENER_SESSION_ID          | *No*         | Session UUIDv4 ID               | `"b7e499fd-f6e1-435c-8ef7-624287ca2bd4"` |
| GREENER_SESSION_DESCRIPTION | *No*         | Session description             | `"My test session"`                      |
| GREENER_SESSION_LABELS      | *No*         | Labels to attach to the session | `"label1=value1,label2"`                 |
| GREENER_SESSION_BAGGAGE     | *No*         | JSON to attach to the session   | `'{"version": "2.0.0"}'`                 |

## Query Language

### Basics
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

### Supported identifiers
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

### Status values
Valid status values: `"pass"`, `"fail"`, `"error"`, `"skip"`

### Modifiers
| Modifier    | Format                        | Description           |
|:------------|:------------------------------|:----------------------|
| offset      | `offset = <number>`           | Skip N results        |
| limit       | `limit = <number>`            | Return max N results  |
| start_date  | `start_date = "YYYY/MM/DD HH:MM:SS"` | Filter from date |
| end_date    | `end_date = "YYYY/MM/DD HH:MM:SS"`   | Filter to date   |

## MCP Server

Greener includes an MCP (Model Context Protocol) server that allows AI agents to query test results.
The MCP server is available at `/api/v1/mcp` and uses OAuth 2.0 for authentication.

### Configuration

Set `GREENER_AUTH_ISSUER` environment variable to your external base URL if it differs from localhost.

## License
This project is licensed under the terms of the [Apache License 2.0](./LICENSE).
