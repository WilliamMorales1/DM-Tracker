# DM Tracker

Local web app for managing D&D games. Includes searchable monster and spell catalogs, initiative tracking, notes, custom maps with markers, quick-reference tables, and optional Ollama-powered chat and lore search.

## Features

- **Monsters** — searchable bestiary with stat blocks, filters, and token images
- **Spells** — searchable spell catalog
- **Initiative** — combat tracker with rounds, hit points, and saved presets
- **Maps** — upload PNG, JPEG, or WebP maps; add labeled markers with notes
- **Notes** — Markdown editor with preview; notes are stored locally
- **Ask Agent** — chat with an Ollama model that can query tracker data
- **Lore Search** — semantic search over imported adventure or book text
- **References** — quick-reference tables for actions, conditions, skills, and ranges
- **Migrations** — SQLite schema is applied automatically at startup

## Requirements

- Go 1.26.4 or newer
- GNU Make
- Modern browser with native ES module support
- Ollama running locally at `http://localhost:11434` for Ask Agent and Lore Search
  - Chat model with tool-calling support, selected in Ask Agent
  - `nomic-embed-text-v2-moe` for lore embeddings

Ollama is optional if chat and lore search are not needed. SQLite is embedded; no separate database service is required.

## Setup

Build and run from repository root:

```sh
make build
make run
```

App listens at `http://localhost:8080`. The database is created at `backend/tracker.db`. Run commands from repository root; server expects `backend/` as its working directory.

For development without building a binary:

```sh
make dev
```

`make build-frontend` checks required static frontend files. `make build-backend` builds `backend/tracker`. Frontend uses plain JavaScript and CSS without a package manager or build step. Vendored Markdown and Monaco files support offline use.

## Importing data

Data imports are optional and require network access. Run commands from repository root:

```sh
cd backend
go run ./cmd/ingest-5etools
```

This imports the default monster sources (`mm,xmm`) and their available images. Add source codes with `-source`, or import every indexed bestiary with `-all`:

```sh
go run ./cmd/ingest-5etools -source mm,xmm,mtf
# or
go run ./cmd/ingest-5etools -all
```

Importing lore replaces existing lore-search chunks. Ollama must be running with the embedding model pulled:

```sh
go run ./cmd/ingest-lore
```

Default lore source is 5etools' Curse of Strahd adventure JSON. Choose another compatible adventure or book JSON by passing a raw URL or GitHub blob URL:

```sh
go run ./cmd/ingest-lore -source 'https://github.com/OWNER/REPO/blob/BRANCH/path/to/source.json'
```

Both import commands accept `-db` to choose a SQLite database path. Run them from `backend/` so default database path resolves correctly.

## Local data

- `backend/tracker.db` — SQLite database, gitignored
- `backend/notes/` — Markdown notes, gitignored
- `backend/maps/` — uploaded map images, gitignored
- `backend/images/` — static app images; user-supplied `.webp` map art is gitignored

Notes created or edited in the app are written to `backend/notes/`. Back up database and local files separately if needed.

## Tests and utilities

Run headless Chromium UI tests (requires Chromium):

```sh
make test-ui
```

Run optional frontend logic tests directly with Node.js:

```sh
node --test frontend/test/*.test.mjs
```

Plot map coordinates for diagnostics:

```sh
make plot MAP=<map-id>
```

Other targets: `make clean` removes the built binary and temporary files.

## Database migrations

Numbered SQL migrations live in `backend/migrations/`. The startup migration runner in `backend/internal/db/` applies pending migrations in transactions and uses the `schema_migrations` table, compatible with databases created by golang-migrate.

Normal startup applies pending migrations automatically. `make migrate` resets and reapplies migrations against `backend/tracker.db`; this is destructive and drops stored tracker data. Back up database before running it.

## Project structure

```text
backend/
  cmd/tracker/       HTTP server, API, and Ollama chat agent
  cmd/migrate/       destructive migration reset utility
  cmd/ingest-5etools/ monster data importer
  cmd/ingest-lore/   adventure/book text and embedding importer
  cmd/plot/          map-coordinate diagnostic tool
  internal/db/       SQLite access and migration runner
  internal/          embedding client
  migrations/        numbered SQLite migrations
  e2e/               Chromium UI tests
  notes/             local Markdown notes (gitignored)
  maps/              uploaded maps (gitignored)
frontend/
  app/               browser modules and panels
  vendor/            checked-in browser dependencies
  test/              optional Node.js logic tests
Makefile             build, run, test, and utility targets
```

## Disclaimer

Unofficial fan tool for Dungeons & Dragons. Dungeons & Dragons and related marks are trademarks of Wizards of the Coast. This project is not affiliated with or endorsed by Wizards of the Coast. Imported 5etools data, adventure text, and user-supplied maps are not included under this project's license.

## License

Code is released under the [MIT License](LICENSE). Vendored libraries under `frontend/vendor/` retain their own licenses. Imported content and user-supplied assets are not covered by this license.
