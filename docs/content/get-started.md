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
go install github.com/cephei8/greener/cmd/greener-admin@main

greener-admin \
    --db-url sqlite:///greener-data/greener.db \
    create-user \
    --username greener \
    --password greener
```

Now, access Greener at http://localhost:8080 in your browser.

## Reporting test results to Greener
Check out [Ecosystem](ecosystem.md) for ways to report test results to Greener.
