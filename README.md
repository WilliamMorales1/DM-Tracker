# DM Tracker

Local D&D game tracker. Stores monsters, spells, notes, initiative presets, and maps in SQLite. Optional chat and lore search use any OpenAI-compatible LLM server (Lemonade, Ollama, llama.cpp, vLLM).

## Features

- Monster and spell catalogs
- Initiative tracker with saved presets
- Maps with uploaded images and markers
- Markdown notes
- Quick references for actions, conditions, skills, and ranges
- Optional LLM chat and lore search

## Requirements

- Go 1.26.4 or newer
- GNU Make
- Modern browser with ES module support
- OpenAI-compatible LLM server for chat and lore search
  - Chat model with tool-calling support
  - `nomic-embed-text-v2-moe` for lore search

LLM server is not needed for other features. SQLite is built into the app.

## LLM server

Without `LLM_BASE_URL`, app probes `GET /models` in this order and uses first server that answers:

1. Lemonade: `http://localhost:13305/api/v1`
2. Ollama: `http://localhost:11434/v1`

Only OpenAI-compatible routes are used: `/models`, `/chat/completions`, `/embeddings`.

| Variable | Default | Purpose |
|---|---|---|
| `LLM_BASE_URL` | auto-detect | Server base URL including `/v1` path, e.g. `http://localhost:8000/v1` |
| `LLM_API_KEY` | unset | Sent as `Authorization: Bearer` when set |
| `LLM_EMBED_MODEL` | `nomic-embed-text-v2-moe` | Embedding model id for lore import and search |

Set up embedding model:

```sh
# Lemonade (registers as user.nomic-embed-text-v2-moe, listed as nomic-embed-text-v2-moe)
lemonade pull user.nomic-embed-text-v2-moe \
  --checkpoint main nomic-ai/nomic-embed-text-v2-moe-GGUF:nomic-embed-text-v2-moe.f16.gguf \
  --recipe llamacpp --label embeddings

# Ollama
ollama pull nomic-embed-text-v2-moe
```

Chat picker lists chat models from server. On Lemonade, only models labeled `chat`; on other servers, all models except ids containing `embed`. Cloud models (`:cloud`, `-cloud`, Lemonade `ollama.*`) are preselected because local CPU inference is slow on agent prompt.

Lore vectors from one embedding model do not match query vectors from another. After changing `LLM_EMBED_MODEL`, rerun `go run ./cmd/ingest-lore`. Same model id on Lemonade and Ollama should give near-identical vectors (both f16 GGUF); reingest if switching servers degrades search.

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

Frontend is TypeScript built with Vite. `make dev` builds it to `frontend/dist/`, which the Go server serves. Node.js 22+ required. For hot reload, run the backend and `cd frontend && npm run dev` (proxies `/api`, `/images`, `/maps` to `localhost:8080`).

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

Importing lore replaces existing lore-search data. Start LLM server and pull embedding model first (see [LLM server](#llm-server)):

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

Frontend tests and type check require Node.js:

```sh
cd frontend && npm test && npm run typecheck
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
backend/internal/llm/      OpenAI-compatible LLM client, server auto-detect
backend/internal/embeddings/ embedding client
backend/migrations/        SQLite migrations
backend/notes/             local notes (ignored by Git)
backend/maps/              uploaded maps (ignored by Git)
frontend/src/              TypeScript source (panels/, lib/, styles/)
frontend/public/           static assets, vendored Monaco editor
frontend/dist/             build output (ignored by Git)
Makefile                   build and run commands
```

## Disclaimer

Unofficial D&D fan tool. Dungeons & Dragons and related marks belong to Wizards of the Coast. This project is not affiliated with Wizards of the Coast. Imported 5etools data, adventure text, and user maps are not included under this project's license.

## License

Code uses the [MIT License](LICENSE). Vendored libraries in `frontend/public/vendor/` keep their own licenses. Imported content and user-supplied assets are not covered by this license.
