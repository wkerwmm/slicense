## License Server (Go)

A simple license management server and CLI written in Go. It provides:

- License CRUD via CLI (add, delete, list) and verification endpoint
- MySQL-backed persistence with audit logs
- Basic account registration/login with JWT issuance
- REST API using `chi` with CORS enabled for `http://localhost:5173`

### Tech Stack

- Go 1.23+ (toolchain 1.24.5)
- MySQL (via `github.com/go-sql-driver/mysql`)
- HTTP router: `github.com/go-chi/chi/v5`
- CLI: `github.com/urfave/cli/v2`
- JWT: `github.com/golang-jwt/jwt/v5`

### Repository Structure

- `main.go`: CLI and HTTP server entry
- `config.yml`: App configuration (MySQL, server port)
- `database/`: DB connector, schema bootstrap, and queries
- `license/`: License domain service and HTTP handlers for verification/audit logs
- `web/`: API routes, auth handlers/services, middleware
- `utils/`: config loader, JWT, password hashing, key generator

### Prerequisites

- Go installed (1.23+)
- MySQL instance accessible and a database created (default: `license_db`)

### Configuration

Configure the app via `config.yml` at repo root:

```yaml
mysql:
  host: localhost
  port: 3306
  user: root
  password: ""
  database: license_db

server:
  port: 8080
```

Notes:

- Tables are created automatically on startup (`licenses`, `Accounts`, `audit_log`).
- JWT secret is currently hardcoded in `utils/jwt.go`. Change it before production.

### Setup

1) Ensure MySQL is running and the configured database exists.
2) Build the binary:

```bash
go build -o license-server
```

3) Run the CLI or start the server (see below).

### CLI Usage

The CLI is embedded in the binary. Examples assume `./license-server` binary in project root.

- Add license:

```bash
./license-server add --key random --product "MyApp" --email user@example.com --name "Jane Doe" --hours 48
```

- Delete license:

```bash
./license-server delete <key> <product>
```

- List licenses for a product:

```bash
./license-server list <product>
```

- Show audit logs (optional limit, default 10):

```bash
./license-server logs 20
```

- Start HTTP server:

```bash
./license-server serve
```

### HTTP Server

When started, the server exposes:

- License endpoints (mounted at root on default mux):
  - `POST /license/verify` — verify license validity
  - `GET  /license/audit-logs` — list recent audit logs

- API router (mounted under `/api`):
  - `GET  /api/ping` — health check
  - `POST /api/auth/register` — create account
  - `POST /api/auth/login` — returns JWT

### Request/Response Examples

- Verify license

Request:

```bash
curl -s -X POST http://localhost:8080/license/verify \
  -H 'Content-Type: application/json' \
  -d '{"key":"ABCD-EFGH-IJKL-MNOP","product":"MyApp"}'
```

Successful response:

```json
{
  "valid": true,
  "key": "ABCD-EFGH-IJKL-MNOP",
  "product": "MyApp",
  "expires_at": null,
  "owner_email": "user@example.com",
  "owner_name": "Jane Doe",
  "is_activated": false
}
```

- Register

```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"jane","email":"user@example.com","password":"Passw0rd!","passwordRepeat":"Passw0rd!"}'
```

- Login (returns JWT):

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Passw0rd!"}'
```

### CORS

- CORS allows requests from `http://localhost:5173`. Adjust in `web/routes.go` if needed.

### Development Tips

- Running directly:

```bash
go run . serve
```

- Database connectivity uses DSN from `config.yml`.
- Passwords are hashed with bcrypt; do not store plaintext.

### Security Considerations

- Replace the hardcoded JWT key in `utils/jwt.go` with a secure secret (env var).
- Use TLS in production and secure your MySQL connection.
- Rate-limit or protect auth endpoints as appropriate.

### License

This project includes basic license management features; see `license/` for domain logic.

