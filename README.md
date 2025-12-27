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
- Attach labels and/or baggage (arbitrary JSON) to test sessions
- Self-contained executable (only requires SQLite/PostgresSQL/MySQL database)
- Small (~27mb executable / compressed Docker image)



Check out [Greener Documentation](https://greener.cephei8.dev) for details.

TODO: demo

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

# create user
go install github.com/cephei8/greener/cmd/greener-admin@main

greener-admin \
    --db-url sqlite:///greener-data/greener.db \
    create-user \
    --username greener \
    --password greener
```

Then open <http://localhost:8080> in your browser.

#### Docker Compose
Check out [compose.yaml](https://github.com/cephei8/greener/blob/main/compose.yaml) for example of using Docker Compose and PostgresSQL.

### Building from source

```bash
# build
npm install
npm run build
go build ./cmd/greener

# run Greener server
mkdir greener-data

./greener --db-url "sqlite:///greener-data/greener.db" --auth-secret "my-secret"

# create user
go install github.com/cephei8/greener/cmd/greener-admin@main

greener-admin \
    --db-url sqlite:///greener-data/greener.db \
    create-user \
    --username greener \
    --password greener
```

Then open <http://localhost:8080> in your browser.

## Reporting test results to Greener

Check out [Ecosystem section](#ecosystem) for ways to report test results to Greener.

For the "hello world" the easiest option may be to use [cephei8/greener-reporter-cli](https://github.com/cephei8/greener-reporter-cli).

## Ecosystem

Plugins/reporters:
- pytest plugin: [cephei8/pytest-greener](https://github.com/cephei8/pytest-greener)
- Jest reporter: [cephei8/jest-greener](https://github.com/cephei8/jest-greener)
- Mocha reporter: [cephei8/mocha-greener](https://github.com/cephei8/mocha-greener)
- Tool to report Go test results: [cephei8/greener-reporter-go](https://github.com/cephei8/greener-reporter-go)
- Tool to report JUnit XML resilts: [cephei8/greener-reporter-junitxml](https://github.com/cephei8/greener-reporter-junitxml)
- CLI tool to report test results: [cephei8/greener-reporter-cli](https://github.com/cephei8/greener-reporter-cli)

Supporting libraries:
- Python library for implementing reporters: [cephei8/greener-reporter-py](https://github.com/cephei8/greener-reporter-py)
- JavaScript library for implementing reporters: [cephei8/greener-reporter-js](https://github.com/cephei8/greener-reporter-js)
- C FFI library for implementing reporters: [cephei8/greener-reporter](https://github.com/cephei8/greener-reporter)

## Contributing
See [CONTRIBUTING.md](./CONTRIBUTING.md).

## License
This project is licensed under the terms of the [Apache License 2.0](./LICENSE).
