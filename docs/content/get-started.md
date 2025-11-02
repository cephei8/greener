# Get Started

## Platform setup
Pull and run the Docker image [cephei8/greener](https://hub.docker.com/r/cephei8/greener).

The following environment variables must be specified:

- `GREENER_DATABASE_URL`: database url, e.g. `postgresql://postgres:qwerty@localhost:5432/postgres?ssl=disable`
- `GREENER_JWT_SECRET` - JWT secret, e.g. `abcdefg1234567`

## Plugin setup

=== "Python: pytest"

    ``` shell
    pip install pytest-greener
    ```

=== "JavaScript: Jest"

    ``` shell
    npm i jest-greener
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

=== "JavaScript: Jest"

    Add the following to `package.json`:
    ``` json
    "scripts": {
      "test": "jest"
    }
    ```

    Run tests:
    ``` shell
    npm test -- --reporters="default" --reporters="jest-greener"
    ```
