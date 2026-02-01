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
