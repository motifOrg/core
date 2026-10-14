# Motif Core

Go API for Motif: Gin, GORM, PostgreSQL and golang-migrate, laid out like Sendin, Dialbot and Lens core.

## Setup

```bash
cp .env.example .env
go mod tidy
go run ./cmd/server
```

On start the server creates the database if it doesn't exist, then applies `internal/migrations`. For peer auth over a Unix socket set `DB_HOST=/var/run/postgresql` and leave `DB_PASSWORD` empty.

Load the sample data (idempotent, every row is flagged `is_sample`):

```bash
./script/bootstrap.bash
```

| Script | What it does |
|--------|--------------|
| `script/bootstrap.bash` | Runs `script/seed.sql` against `DB_NAME` |
| `script/migrate.bash up\|down\|version\|force <v>` | golang-migrate CLI wrapper |
| `script/migrations.bash <name>` | Creates the next numbered up/down migration pair |
| `script/swagger.bash` | Generates Swagger docs from the controller annotations (needs `swag`) |

```bash
go test ./...
go vet ./...
```

## Layout

```
cmd/server/            entrypoint
api/router/            routes and CORS
api/controller/        handlers, request types, validation messages
internal/config/       env config
internal/database/     ensure, connect, migrate
internal/migrations/   SQL migrations
internal/model/        GORM models and enums
internal/repository/   queries
internal/service/      research and Gmail
pkg/csv/               company CSV parser
pkg/gmail/             Gmail API client (service account)
pkg/log/               logger
pkg/openai/            OpenAI client, enrichment agent and research tools
script/                bootstrap, migrate, seed, swagger
```

## API

Base path `/api/v1`. Responses are `{"data": ...}` (lists add `total`, `page`, `count`) or `{"error": "..."}`. Validation errors name the field, for example `sender_email must be a valid email address; subject is required`.

| Method | Path | |
|--------|------|-|
| GET | `/health` | |
| GET | `/integrations` | Gmail and research status, including missing env vars |
| POST | `/integrations/gmail/sync?campaign_id=` | Pulls messages from the campaign `gmail_label` (last 7 days; `INBOX` when empty) |
| GET, POST | `/campaigns` | `?status=` (`all` includes archived) `?q=` |
| GET, PATCH, DELETE | `/campaigns/:id` | |
| GET | `/campaigns/:id/stats` | |
| GET, PUT | `/campaigns/:id/icp` | |
| POST | `/campaigns/:id/companies/import` | CSV; returns inserted / skipped / failed with row errors |
| GET, POST | `/companies` | `?campaign_id=` `?q=` `?sort=` `?priority=` `?research_status=` `?evaluation_status=` `?page=` |
| GET, PATCH, DELETE | `/companies/:id` | |
| POST, DELETE | `/companies/:id/findings[/:findingId]` | |
| POST | `/companies/:id/research` | 503 until a research provider is configured |
| GET, POST | `/companies/:id/messages` | Research chat; stored as notes without a provider |
| GET, POST | `/inbound` | Returns per-status `counts` |
| GET, PATCH | `/inbound/:id` | |
| GET, POST | `/drafts` | `?company_id=` or `?opportunity_id=` |
| PATCH, DELETE | `/drafts/:id` | Editing a reviewed draft resets it to `draft` |

## Integration seams

Both live in `internal/service` and are nil today, so the API reports them as not configured instead of faking results.

- `ResearchProvider`: `Research(company, icp)` returns findings, `Ask(company, history, question)` returns a reply. This is where web search, LLM orchestration or MCP tools would go.
- Gmail: uses a Google Workspace service account with domain-wide delegation. In Google Cloud, create a service account and download its JSON key. In Workspace admin (Security → API controls → Domain-wide delegation), add the service account client ID with scope `https://www.googleapis.com/auth/gmail.readonly`. Set `gmail_label` per campaign in the app; when empty, sync uses `INBOX`.

```env
GMAIL_CREDENTIALS=/path/to/service-account.json
GMAIL_USER=deals@yourfirm.com
```

There is no send path anywhere in the API.
