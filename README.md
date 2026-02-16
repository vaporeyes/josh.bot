# josh.bot

A personal API-first platform for Josh, accessible at [josh.bot](https://josh.bot). Built as a cloud-native backend that powers [k8-one](https://k8-one.com) (Josh's AI agent on Slack) with real-time status, project tracking, link management, notes, TILs, activity logging, fitness metrics, and development memory (claude-mem observations, summaries, and prompts).

## Architecture

Built with **Hexagonal Architecture** (Ports and Adapters) in Go. The core domain logic is isolated from external concerns, making the system modular, testable, and easy to extend.

```
cmd/
  api/            Local dev server (mock data, no auth)
  lambda/         Production entrypoint (DynamoDB, API key auth)
  import-lifts/   CLI tool for importing Strong app workout CSV exports
  export-links/   CLI tool for exporting links with tag/date filters (JSON or URL-only)
  sync-mem/       CLI tool for syncing claude-mem SQLite to DynamoDB
internal/
  domain/         Core types, service interfaces, and calculation helpers
  adapters/
    dynamodb/     DynamoDB-backed service implementation
    lambda/       API Gateway event routing
    http/         HTTP handlers for local dev
    mock/         In-memory service for testing
scripts/          Seed scripts and DynamoDB table creation
terraform/        Infrastructure as code
```

### Data Model

Uses DynamoDB **single-table design** with the `id` partition key and prefixed keys:

| Prefix | Example ID | Resource |
|--------|-----------|----------|
| `status` | `status` | Bot owner status (singleton) |
| `project#` | `project#modular-aws-backend` | Projects by slug |
| `link#` | `link#a1b2c3d4e5f6` | Links by SHA256 URL hash |
| `note#` | `note#a1b2c3d4e5f6a1b2` | Notes (random ID) |
| `til#` | `til#a1b2c3d4e5f6a1b2` | TIL entries (random ID) |
| `log#` | `log#a1b2c3d4e5f6a1b2` | Activity log entries (random ID) |

Link IDs are derived from the URL via SHA256, giving automatic deduplication -- saving the same URL twice updates the existing entry. Notes, TILs, and log entries use random 8-byte hex IDs.

Lift/workout data lives in a separate `josh-bot-lifts` table with a `date-index` GSI for time-range queries. Lift IDs are deterministic (date + exercise + set order) making CSV re-imports idempotent.

Development memory (claude-mem) data lives in a separate `josh-bot-mem` table with a `type-index` GSI (partition key: `type`, sort key: `created_at_epoch`):

| Prefix | Example ID | Resource |
|--------|-----------|----------|
| `obs#` | `obs#142` | Development observations (decisions, features, bugs) |
| `summary#` | `summary#85` | Session summaries (what was investigated/completed) |
| `prompt#` | `prompt#301` | User prompts from coding sessions |

Data is synced from the local claude-mem SQLite database using `cmd/sync-mem`.

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
curl http://localhost:8080/v1/metrics
curl http://localhost:8080/v1/projects
curl http://localhost:8080/v1/links
curl http://localhost:8080/v1/notes
curl http://localhost:8080/v1/til
curl http://localhost:8080/v1/log
curl http://localhost:8080/v1/mem/observations
curl http://localhost:8080/v1/mem/summaries
curl http://localhost:8080/v1/mem/prompts
curl http://localhost:8080/v1/mem/stats
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

All endpoints return JSON. Write endpoints require an `x-api-key` header. `GET /v1/status` and `GET /v1/metrics` are the only public (unauthenticated) routes.

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

### Metrics

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/metrics` | No | Get fitness and performance metrics dashboard |

```bash
curl https://api.josh.bot/v1/metrics
```

**Response shape:**

```json
{
  "timestamp": "2026-02-15T12:00:00Z",
  "human": {
    "focus": "Powerlifting / Hypertrophy",
    "weekly_tonnage_lbs": 42500,
    "estimated_1rm": { "deadlift": 525, "squat": 455, "bench": 315 },
    "last_workout": {
      "date": "2026-02-14",
      "name": "Pull Day",
      "exercises": ["Deadlift (Barbell)", "Barbell Row"],
      "sets": 18,
      "tonnage_lbs": 12500
    }
  },
  "dev": {
    "total_observations": 150,
    "total_summaries": 30,
    "total_prompts": 75,
    "by_type": { "decision": 45, "feature": 60, "bugfix": 25 },
    "by_project": { "josh.bot": 120, "other": 30 }
  }
}
```

### Notes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/notes` | Yes | List all notes (optional `?tag=` filter) |
| POST | `/v1/notes` | Yes | Create a note |
| GET | `/v1/notes/{id}` | Yes | Get a note by ID |
| PUT | `/v1/notes/{id}` | Yes | Partial update (allowed fields: `title`, `body`, `tags`) |
| DELETE | `/v1/notes/{id}` | Yes | Delete a note |

```bash
# Create a note
curl -X POST https://api.josh.bot/v1/notes \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"title": "API design thoughts", "body": "Keep endpoints RESTful...", "tags": ["architecture"]}'

# List notes filtered by tag
curl -H "x-api-key: <key>" "https://api.josh.bot/v1/notes?tag=architecture"
```

### TIL (Today I Learned)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/til` | Yes | List all TILs (optional `?tag=` filter) |
| POST | `/v1/til` | Yes | Create a TIL entry |
| GET | `/v1/til/{id}` | Yes | Get a TIL by ID |
| PUT | `/v1/til/{id}` | Yes | Partial update (allowed fields: `title`, `body`, `tags`) |
| DELETE | `/v1/til/{id}` | Yes | Delete a TIL |

```bash
# Create a TIL
curl -X POST https://api.josh.bot/v1/til \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"title": "Go slices grow by 2x", "body": "When capacity is exceeded, Go doubles the backing array", "tags": ["go"]}'

# List TILs filtered by tag
curl -H "x-api-key: <key>" "https://api.josh.bot/v1/til?tag=go"
```

### Activity Log

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/log` | Yes | List all entries (optional `?tag=` filter) |
| POST | `/v1/log` | Yes | Create a log entry |
| GET | `/v1/log/{id}` | Yes | Get a log entry by ID |
| PUT | `/v1/log/{id}` | Yes | Partial update (allowed fields: `message`, `tags`) |
| DELETE | `/v1/log/{id}` | Yes | Delete a log entry |

```bash
# Log an activity
curl -X POST https://api.josh.bot/v1/log \
  -H "x-api-key: <key>" -H "Content-Type: application/json" \
  -d '{"message": "deployed josh-bot v1.3", "tags": ["deploy", "infra"]}'

# List recent activity filtered by tag
curl -H "x-api-key: <key>" "https://api.josh.bot/v1/log?tag=deploy"
```

### Development Memory (claude-mem)

Read-only access to development observations, session summaries, and prompts synced from the claude-mem database.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/mem/observations` | Yes | List observations (optional `?type=` and `?project=` filters) |
| GET | `/v1/mem/observations/{id}` | Yes | Get a single observation |
| GET | `/v1/mem/summaries` | Yes | List session summaries (optional `?project=` filter) |
| GET | `/v1/mem/summaries/{id}` | Yes | Get a single summary |
| GET | `/v1/mem/prompts` | Yes | List user prompts |
| GET | `/v1/mem/prompts/{id}` | Yes | Get a single prompt |
| GET | `/v1/mem/stats` | Yes | Aggregate counts by type and project |

```bash
# List all observations
curl -H "x-api-key: <key>" https://api.josh.bot/v1/mem/observations

# Filter observations by type
curl -H "x-api-key: <key>" "https://api.josh.bot/v1/mem/observations?type=decision"

# Filter observations by project
curl -H "x-api-key: <key>" "https://api.josh.bot/v1/mem/observations?type=feature&project=josh.bot"

# Get session summaries
curl -H "x-api-key: <key>" https://api.josh.bot/v1/mem/summaries

# Get aggregate stats
curl -H "x-api-key: <key>" https://api.josh.bot/v1/mem/stats
```

### CLI Tools

#### export-links

Export links from DynamoDB with filtering for ArchiveBox or other tools.

```bash
# Export all links as JSON
go run cmd/export-links/main.go

# Export URLs only (for piping to ArchiveBox)
go run cmd/export-links/main.go --urls-only

# Filter by tag
go run cmd/export-links/main.go --tag=go

# Filter by date range
go run cmd/export-links/main.go --since=2026-01-01
```

#### sync-mem

Sync claude-mem SQLite data to the `josh-bot-mem` DynamoDB table.

```bash
# Full sync (all observations, summaries, prompts)
go run cmd/sync-mem/main.go

# Dry run (show what would be synced)
go run cmd/sync-mem/main.go --dry-run

# Incremental sync (only records after a timestamp)
go run cmd/sync-mem/main.go --since="2026-02-01 00:00:00"

# Filter by project
go run cmd/sync-mem/main.go --project=josh.bot
```

## Infrastructure

Managed with Terraform in the `terraform/` directory:

| Resource | Purpose |
|----------|---------|
| **AWS Lambda** (`provided.al2023`, ARM64) | Runs the Go API |
| **API Gateway** (HTTP API) | Routes requests to Lambda (10 rps / 20 burst rate limit, default endpoint disabled) |
| **DynamoDB** `josh-bot-data` (PAY_PER_REQUEST) | Single-table store for status, projects, links, notes, TILs, log entries |
| **DynamoDB** `josh-bot-lifts` (PAY_PER_REQUEST) | Workout/lift data with `date-index` GSI |
| **DynamoDB** `josh-bot-mem` (PAY_PER_REQUEST) | Claude-mem data (observations, summaries, prompts) with `type-index` GSI |
| **ACM + Route53** | Custom domain (`api.josh.bot`) with TLS 1.2 |
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
- **Rate limiting**: 10 requests/sec with 20 burst at the API Gateway stage level
- **Custom domain only**: Default API Gateway execute-api endpoint is disabled

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
