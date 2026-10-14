# Motif Core

Go API for Motif: Gin, GORM, PostgreSQL and golang-migrate, with OpenAI for research and Gmail IMAP for inbound. Laid out like Sendin, Dialbot and Lens core.

## Setup

```bash
cp .env.example .env
go mod download
go run ./cmd/server
```

On start the server creates the database if it doesn't exist, then applies `internal/migrations`. For peer auth over a Unix socket, set `DB_HOST=/var/run/postgresql` and leave `DB_PASSWORD` empty.

| Variable | Needed for | Notes |
|----------|-----------|-------|
| `PORT` | Server | Default `8000`. Cloud Run sets it |
| `APP_ORIGIN` | CORS | Default `http://localhost:3000` |
| `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE` | Postgres | |
| `OPENAI_API_KEY` | Research, chat, inbound scoring, outreach drafts | Without it, `/integrations` reports research as `setup_required` and those routes return 503 or fall back to templates |
| `GMAIL_USER`, `GMAIL_APP_PASSWORD` | Gmail sync | See [Gmail](#gmail). Spaces in the app password are stripped |

Data scripts:

| Script | What it does |
|--------|--------------|
| `script/bootstrap.bash` | Runs `script/seed.sql`: fictional sample data, idempotent, every row flagged `is_sample` |
| `script/demo.sql` | Creates the "Bootstrapped B2B SaaS" campaign to pair with `../demo-leads.csv` |
| `script/migrate.bash up\|down\|version\|force <v>` | golang-migrate CLI wrapper |
| `script/migrations.bash <name>` | Creates the next numbered up/down migration pair |
| `script/swagger.bash` | Generates Swagger docs from the controller annotations (needs `swag`) |

Docker: `docker build -t motif-api .` produces a distroless image with the binary and `internal/migrations`. `../docker-compose.yaml` runs it with Postgres and the app.

Checks: `go build ./... && go vet ./...`

## Layout

```
cmd/server/            entrypoint; configures the OpenAI and Gmail clients
api/router/            routes and CORS
api/controller/        handlers, request types, validation messages
internal/config/       env config
internal/database/     ensure, connect, migrate
internal/migrations/   SQL migrations (the schema lives here, not in GORM)
internal/model/        GORM models and enums
internal/repository/   queries
internal/service/      research and Gmail services used by the controllers
pkg/csv/               company CSV parser
pkg/gmail/             IMAP client and the free-mail domain list
pkg/log/               logger
pkg/openai/            OpenAI client, agent loop, research tools and prompts
script/                bootstrap, demo, migrate, seed, swagger
```

## API

Base path `/api/v1`. Responses are `{"data": ...}` (lists add `total`, `page` and `count`) or `{"error": "..."}`. Validation errors name the field, for example `sender_email must be a valid email address; subject is required`.

| Method | Path | |
|--------|------|-|
| GET | `/health` | |
| GET | `/integrations` | Gmail and research status, including missing env vars |
| POST | `/integrations/gmail/sync?campaign_id=` | Pulls the campaign's label, then scores new deals and enriches new leads in the background |
| GET, POST | `/campaigns` | `?status=` (`all` includes archived) `?q=` |
| GET, PATCH, DELETE | `/campaigns/:id` | PATCH accepts `gmail_label` |
| GET | `/campaigns/:id/stats` | |
| GET, PUT | `/campaigns/:id/icp` | |
| POST | `/campaigns/:id/companies/import` | CSV; returns inserted / skipped / failed with row errors; enriches new rows in the background |
| GET, POST | `/companies` | `?campaign_id=` `?q=` `?sort=` `?priority=` `?research_status=` `?evaluation_status=` `?page=` `?limit=` |
| GET | `/companies/export` | Same filters as the list, default `sort=priority`; `text/csv` with fit, reasons, risks and source URLs |
| GET, PATCH, DELETE | `/companies/:id` | |
| POST, DELETE | `/companies/:id/findings[/:findingId]` | Manual findings |
| POST | `/companies/:id/research` | Runs the agent and waits for it; 503 without `OPENAI_API_KEY` |
| GET, POST | `/companies/:id/messages` | Research chat; the agent replies when configured, otherwise messages are saved as notes |
| GET, POST | `/inbound` | List returns per-status `counts`; POST logs a deal and scores it in the background |
| GET, PATCH | `/inbound/:id` | |
| POST | `/inbound/:id/evaluate` | Scores the deal now and returns it; 503 without `OPENAI_API_KEY` |
| GET, POST | `/drafts` | `?company_id=` or `?opportunity_id=`. Outreach drafts are written by AI; reply drafts use a template |
| PATCH, DELETE | `/drafts/:id` | Editing a reviewed draft resets it to `draft` |

There is no send path anywhere in the API.

## Research agent

`pkg/openai` uses the OpenAI Responses API with strict JSON schemas. The agent loop (`mcp.go`) gives the model two function tools, and the model calls them only when the data is missing:

- `revenue_research` runs a web search for revenue and headcount, with sources and a confidence level.
- `news_research` runs a web search for news from the last 24 months, tagged as fit, risk or finding, each with a source URL.

| Call | Input | Writes |
|------|-------|--------|
| `Enrich` | Company and its findings, ICP | Company `fit_score`, `fit_summary`, `priority`, revenue and headcount when found, findings marked verified when they have a source |
| `Evaluate` | Inbound email (sender, subject, body), the linked company if any, ICP | Opportunity `fit_score`, `fit_summary`, `priority`. Criteria the email doesn't mention are reported as unknown, not as misses. No tools run without a company domain |
| `Ask` | Company, ICP, chat history, question | Chat reply with cited sources |
| `Email` (`emailAI.go`) | Company and its findings, campaign name, description, thesis and ICP | Outreach draft subject and body; no web search |

`internal/service/research.go` wraps these calls. It sets `research_status` and resets it if a run fails, and it saves results through the repositories. Controllers only talk to the service, so the provider can be swapped or moved behind a queue without changing the routes.

## Gmail

Sync reads one mailbox over IMAP (`imap.gmail.com:993`) with an app password, which doesn't expire unless revoked.

1. Turn on 2-step verification for the Google account.
2. Create an app password at [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords).
3. Set the two variables and restart the API:

```env
GMAIL_USER=you@gmail.com
GMAIL_APP_PASSWORD=abcdefghijklmnop
```

Each campaign's `gmail_label` is opened as an IMAP folder. Labels keep their Gmail name, and nested ones are written `Parent/Child`. When the label is empty, sync uses `INBOX`.

Sync opens the folder read-only, so nothing is marked as read. It takes the last 7 days, up to 100 messages, and saves the first plain-text part of each body, up to 4,000 characters. Messages are deduplicated by Message-ID.

A business sender domain becomes a company lead in the campaign. Free-mail domains (`pkg/gmail/freemail.go`) are skipped.
