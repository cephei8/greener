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
  -p 5096:5096 \
  -e GREENER_DATABASE_URL=sqlite:///data/db.sqlite \
  -e GREENER_AUTH_SECRET=my-secret \
  cephei8/greener:latest
```

Create user (in a separate shell):
``` shell
go install github.com/cephei8/greener/util/greener-admin-cli@latest

$GOPATH/bin/greener-admin-cli \
  --db-url sqlite:///greener-data/db.sqlite \
  create-user --username greener --password greener
```

Now, access Greener at http://localhost:5096 in the browser.

## Plugin setup

=== "Python: pytest"

    ``` shell
    pip install pytest-greener
    ```

#### Configure plugin

``` shell
export GREENER_INGRESS_ENDPOINT=http://localhost:5096
export GREENER_INGRESS_API_KEY=<api-key>
```
Create API key on Greener API Keys page.

#### Run tests with plugin enabled
=== "Python: pytest"

    ``` shell
    pytest --greener
    ```
