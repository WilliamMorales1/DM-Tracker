# DM Tracker

Local D&D game tracker. Stores monsters, spells, notes, initiative presets, and maps in SQLite. Optional features use Ollama for chat and lore search.

## Features

- Monster and spell catalogs
- Initiative tracker with saved presets
- Maps with uploaded images and markers
- Markdown notes
- Quick references for actions, conditions, skills, and ranges
- Optional Ollama chat and lore search

## Requirements

- Go 1.26.4 or newer
- GNU Make
- Modern browser with ES module support
- Ollama at `http://localhost:11434` for chat and lore search
  - Chat model with tool-calling support
  - `nomic-embed-text-v2-moe` for lore search

Ollama is not needed for other features. SQLite is built into the app.

## Run

From repository root:

```sh
make build
make run
```

Open `http://localhost:8080`. Database file is `backend/tracker.db`.

For development, run without building binary:

```sh
make dev
```

Frontend uses browser JavaScript and CSS. No frontend package manager or build step is needed. Browser dependencies are in `frontend/vendor/`.

## Import data

Imports need network access. Run commands from `backend/`.

Import default monster sources (`mm,xmm`):

```sh
cd backend
go run ./cmd/ingest-5etools
```

Add sources with `-source`, or import every indexed bestiary with `-all`:

```sh
go run ./cmd/ingest-5etools -source mm,xmm,mtf
go run ./cmd/ingest-5etools -all
```

Importing lore replaces existing lore-search data. Start Ollama and pull the embedding model first:

```sh
go run ./cmd/ingest-lore
```

Default lore source is 5etools' Curse of Strahd JSON. Use `-source` to choose another adventure or book JSON. Raw URLs and GitHub blob URLs are supported:

```sh
go run ./cmd/ingest-lore -source 'https://github.com/OWNER/REPO/blob/BRANCH/path/to/source.json'
```

Both import commands accept `-db` to specify database path. Run commands from `backend/` to use default path `tracker.db`.

## Local files

- `backend/tracker.db`: database, ignored by Git
- `backend/notes/`: notes, ignored by Git
- `backend/maps/`: uploaded maps, ignored by Git
- `backend/images/`: static app images; user-supplied WebP maps are ignored by Git

Notes made in app are saved in `backend/notes/`. Back up database and local files separately.

## Tests and tools

UI tests require Chromium:

```sh
make test-ui
```

Frontend logic tests require Node.js:

```sh
node --test frontend/test/*.test.mjs
```

Plot map coordinates:

```sh
make plot MAP=<map-id>
```

`make clean` removes built binary and temporary files.

## Migrations

SQLite migrations are in `backend/migrations/`. App applies pending migrations on startup.

`make migrate` resets and reapplies migrations against `backend/tracker.db`. This deletes stored tracker data. Back up database first.

## Project files

```text
backend/cmd/tracker/       HTTP server, API, chat agent
backend/cmd/migrate/       migration reset utility
backend/cmd/ingest-5etools/ monster importer
backend/cmd/ingest-lore/   lore importer
backend/cmd/plot/          map coordinate tool
backend/internal/db/       database access and migrations
backend/internal/          embedding client
backend/migrations/        SQLite migrations
backend/e2e/               Chromium UI tests
backend/notes/             local notes (ignored by Git)
backend/maps/              uploaded maps (ignored by Git)
frontend/app/              browser modules and panels
frontend/vendor/           browser dependencies
frontend/test/             Node.js tests
Makefile                   build and run commands
```

## Disclaimer

Unofficial D&D fan tool. Dungeons & Dragons and related marks belong to Wizards of the Coast. This project is not affiliated with Wizards of the Coast. Imported 5etools data, adventure text, and user maps are not included under this project's license.

## License

Code uses the [MIT License](LICENSE). Vendored libraries in `frontend/vendor/` keep their own licenses. Imported content and user-supplied assets are not covered by this license.
