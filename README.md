# josh.bot

A personal API-first platform for Josh, accessible at [josh.bot](https://josh.bot). Built as a cloud-native backend that powers [k8-one](https://k8-one.com) (Josh's AI agent on Slack) with real-time status, project tracking, and link management.

## Architecture

Built with **Hexagonal Architecture** (Ports and Adapters) in Go. The core domain logic is isolated from external concerns, making the system modular, testable, and easy to extend.

```
cmd/
  api/          Local dev server (mock data, no auth)
  lambda/       Production entrypoint (DynamoDB, API key auth)
internal/
  domain/       Core types and service interface
  adapters/
    dynamodb/   DynamoDB-backed service implementation
    lambda/     API Gateway event routing
    http/       HTTP handlers for local dev
    mock/       In-memory service for testing
scripts/        Seed scripts for DynamoDB data
terraform/      Infrastructure as code
```

### Data Model

Uses DynamoDB **single-table design** with the `id` partition key and prefixed keys:

| Prefix | Example ID | Resource |
|--------|-----------|----------|
| `status` | `status` | Bot owner status (singleton) |
| `project#` | `project#modular-aws-backend` | Projects by slug |
| `link#` | `link#a1b2c3d4e5f6` | Links by SHA256 URL hash |

Link IDs are derived from the URL via SHA256, giving automatic deduplication -- saving the same URL twice updates the existing entry.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) 1.24+
- [Task](https://taskfile.dev) (task runner)
- [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) 1.0+
- [golangci-lint](https://golangci-lint.run/welcome/install-locally/)
- [pre-commit](https://pre-commit.com/#install) (optional)
- An [AWS account](https://aws.amazon.com/) for deployment

### Local Development

The local server uses mock data with no API key required.

```bash
git clone https://github.com/vaporeyes/josh.bot.git
cd josh.bot

# Install pre-commit hooks (optional but recommended)
pre-commit install

# Run the local server
go run cmd/api/main.go
```

The server starts on `http://localhost:8080`. Test it:

```bash
curl http://localhost:8080/v1/status
curl http://localhost:8080/v1/projects
curl http://localhost:8080/v1/links
```

### Task Runner

Common tasks are managed via [Taskfile](https://taskfile.dev):

```bash
task --list         # See all available tasks
task check          # Run all checks (fmt, vet, lint, test)
task test           # Run Go tests
task lint           # Run golangci-lint
task build          # Build Lambda binary
task package        # Build + zip for deployment
task deploy         # Full deploy (check, package, terraform apply)
task seed           # Seed DynamoDB with initial data
```

## API Reference

All endpoints return JSON. Write endpoints require an `x-api-key` header. `GET /v1/status` is the only public (unauthenticated) route.

### Status

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/status` | No | Get bot owner status |
| PUT | `/v1/status` | Yes | Partial update (allowed fields: `current_activity`, `location`, `availability`, `status`, `bio`, `title`, `interests`, `links`) |

```bash
# Get status
curl https://api.josh.bot/v1/status

# Update status
curl -X PUT https://api.josh.bot/v1/status \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"current_activity": "Building APIs", "availability": "Busy"}'
```

**Response shape:**

```json
{
  "name": "Josh Duncan",
  "title": "Platform Engineer",
  "bio": "...",
  "current_activity": "Building APIs",
  "location": "Clarksville, TN",
  "availability": "Open to interesting projects",
  "status": "ok",
  "links": { "github": "https://github.com/vaporeyes", "..." : "..." },
  "interests": ["Go", "AWS", "..."],
  "updated_at": "2026-02-14T12:00:00Z"
}
```

### Projects

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/projects` | Yes | List all projects |
| POST | `/v1/projects` | Yes | Create a project |
| GET | `/v1/projects/{slug}` | Yes | Get a project by slug |
| PUT | `/v1/projects/{slug}` | Yes | Partial update (allowed fields: `name`, `stack`, `description`, `url`, `status`) |
| DELETE | `/v1/projects/{slug}` | Yes | Delete a project |

```bash
# List projects
curl -H "x-api-key: <key>" https://api.josh.bot/v1/projects

# Create a project
curl -X POST https://api.josh.bot/v1/projects \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"slug":"my-project","name":"My Project","stack":"Go","description":"A thing","url":"https://github.com/vaporeyes/my-project","status":"active"}'

# Get a single project
curl -H "x-api-key: <key>" https://api.josh.bot/v1/projects/my-project

# Update a project
curl -X PUT https://api.josh.bot/v1/projects/my-project \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"status": "archived"}'

# Delete a project
curl -X DELETE https://api.josh.bot/v1/projects/my-project \
  -H "x-api-key: <key>"
```

### Links / Bookmarks

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/links` | Yes | List all links (optional `?tag=` filter) |
| POST | `/v1/links` | Yes | Save a link (idempotent via URL hash) |
| GET | `/v1/links/{id}` | Yes | Get a link by ID |
| PUT | `/v1/links/{id}` | Yes | Partial update (allowed fields: `title`, `tags`) |
| DELETE | `/v1/links/{id}` | Yes | Delete a link |

```bash
# List all links
curl -H "x-api-key: <key>" https://api.josh.bot/v1/links

# Filter by tag
curl -H "x-api-key: <key>" "https://api.josh.bot/v1/links?tag=aws"

# Save a link (ID auto-generated from URL hash)
curl -X POST https://api.josh.bot/v1/links \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev/blog/","title":"The Go Blog","tags":["go","programming"]}'

# Get a specific link
curl -H "x-api-key: <key>" https://api.josh.bot/v1/links/a1b2c3d4e5f6

# Update a link's tags
curl -X PUT https://api.josh.bot/v1/links/a1b2c3d4e5f6 \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"tags": ["go", "programming", "blog"]}'

# Delete a link
curl -X DELETE https://api.josh.bot/v1/links/a1b2c3d4e5f6 \
  -H "x-api-key: <key>"
```

## Infrastructure

Managed with Terraform in the `terraform/` directory:

| Resource | Purpose |
|----------|---------|
| **AWS Lambda** (`provided.al2023`, ARM64) | Runs the Go API |
| **API Gateway** (HTTP API) | Routes requests to Lambda |
| **DynamoDB** (PAY_PER_REQUEST) | Single-table data store |
| **ACM + Route53** | Custom domain (`api.josh.bot`) with TLS |
| **SSM Parameter Store** | Stores the generated API key |
| **IAM** | Lambda execution role with scoped DynamoDB permissions |
| **S3** | Terraform state backend with native locking |

## CI/CD

Defined in `.github/workflows/cicd.yaml`, triggered on push to `main`:

1. **Check job** -- gofmt, go vet, golangci-lint, go test, terraform fmt, terraform validate
2. **Build and Deploy job** (requires check to pass) -- builds the Lambda binary, authenticates via OIDC, and runs `terraform apply`

Required GitHub secrets: `AWS_ACCOUNT_ID`, `TERRAFORM_BUCKET`

## Code Quality

- **Pre-commit hooks**: `go-fmt`, `go-build`, `go-unit-tests`, `golangci-lint`, `terraform_fmt`, `terraform_validate`
- **TDD**: All features built test-first with mocked DynamoDB client
- **Field allowlists**: Write endpoints only accept known fields, preventing arbitrary data injection
- **API key auth**: All write endpoints (and most reads) require `x-api-key` header

## Seeding Data

After initial deployment, seed the DynamoDB table:

```bash
task seed                    # Seeds status + projects
./scripts/seed-links.sh      # Seeds example bookmarks
```

Or individually:

```bash
./scripts/seed-status.sh     # Bot owner profile
./scripts/seed-projects.sh   # Example projects
./scripts/seed-links.sh      # Example bookmarks
```
