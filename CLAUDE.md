# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

Multi-language client library monorepo for the Lab Nocturne Images API (`https://images.labnocturne.com`). Provides installable client packages in Python, Go, Ruby, JavaScript, and PHP, plus Claude Code skills and a ChatGPT Action integration.

The API server lives elsewhere (`jjenkins/labnocturne`). This repo contains only client-side code and skill definitions.

## Repository Layout

- `python/` — Python client (`labnocturne` package, uses `requests`)
- `go/` — Go client (`labnocturne` package, stdlib only)
- `javascript/` — Node.js client (ESM, uses `form-data`)
- `ruby/` — Ruby client (`labnocturne` gem, stdlib `net/http`)
- `php/` — PHP client (uses `ext-curl`, published to Packagist as `labnocturne/image-client`)
- `examples/curl/` — curl-based usage examples
- `skills/` — Claude Code skill definitions (symlinks to `.agents/skills/`)
- `integrations/chatgpt-action/` — ChatGPT GPT Action with OpenAPI schema
- `.claude/commands/test-example.md` — Custom command to test language examples

## Build & Test Commands

**Python:**
```bash
pip install ./python                          # install locally
pip install ./python[dev]                     # install with dev deps (pytest, black, flake8, mypy)
cd python && pytest                           # run tests
```

**Go:**
```bash
cd go && go build ./...                       # build
cd go && go test ./...                        # test
```

**JavaScript:**
```bash
cd javascript && npm install                  # install deps
cd javascript && npm test                     # run tests (node --test)
```

**Ruby:**
```bash
cd ruby && bundle install                     # install deps
```

**PHP:**
```bash
composer install                              # install from root composer.json
```

## Client Library Conventions

All five clients implement the same API surface with language-idiomatic naming:

| Operation | Endpoint | Python | Go | JS | Ruby | PHP |
|-----------|----------|--------|----|----|------|-----|
| Generate key | `GET /key` | `generate_test_key()` (static) | `GenerateTestKey()` | `generateTestKey()` (static) | `generate_test_key` (class) | `generateTestKey()` (static) |
| Upload | `POST /upload` | `upload(path)` | `Upload(path)` | `upload(path)` | `upload(path)` | `upload(path)` |
| List files | `GET /files` | `list_files()` | `ListFiles()` | `listFiles()` | `list_files()` | `listFiles()` |
| Stats | `GET /stats` | `get_stats()` | `GetStats()` | `getStats()` | `get_stats` | `getStats()` |
| Delete | `DELETE /i/:id` | `delete_file(id)` | `DeleteFile(id)` | `deleteFile(id)` | `delete_file(id)` | `deleteFile(id)` |

All clients default `base_url` to `https://images.labnocturne.com` and authenticate via `Authorization: Bearer <key>` header. Upload uses `multipart/form-data`.

## Skills (Claude Code)

Skills are in `skills/` as symlinks. Each skill is a `SKILL.md` with YAML frontmatter (name, description, metadata tags). The `image-*` prefixed skills are the canonical set:

- `image-upload` — Upload an image, get CDN URL
- `image-files` — List uploaded files with pagination
- `image-stats` — View storage usage
- `image-delete` — Soft-delete an image
- `image-key` — Generate a test API key

Skills auto-detect `$LABNOCTURNE_API_KEY` env var. If unset, they generate a temporary test key (7-day retention, 10MB limit). Skills use `$LABNOCTURNE_BASE_URL` if set.

Shared auth logic (API key resolution, base URL, common error codes) lives in `.agents/skills/references/auth.md` and is referenced by all skills.

The non-prefixed skills (`upload`, `files`, `stats`, `delete`, `generate-key`) are legacy aliases that redirect to their `image-*` counterparts.

## API Key Types

- **Test keys** (`ln_test_*`): free, no signup, 10MB file limit, 7-day retention
- **Live keys** (`ln_live_*`): paid, 100MB file limit, permanent storage

## Style

- Keep client code simple and beginner-friendly with minimal dependencies
- Use real, working code — not pseudocode
- Follow each language's idiomatic conventions
- Keep dependencies minimal (Go and Ruby clients use only stdlib)
- Test examples against the live API before committing (`/test-example <language>`)
