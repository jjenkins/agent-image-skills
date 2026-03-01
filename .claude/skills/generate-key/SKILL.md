---
name: generate-key
description: "Generate a test API key for Lab Nocturne. Triggered by /generate-key or naturally (e.g. 'get me an API key')."
---

# Generate Test API Key

Generate a test API key for the Lab Nocturne Images API.

## Invocation

`/generate-key` — also triggered naturally (e.g. "get me an API key", "generate a key").

## Instructions

Follow these steps exactly:

### 1. Resolve the base URL

- Use `$LABNOCTURNE_BASE_URL` if set, otherwise default to `https://images.labnocturne.com`.

### 2. Generate the key

Run:
```
curl -s <base_url>/key
```

### 3. Handle the response

**Success** — the response body is:
```json
{
  "api_key": "ln_test_...",
  "type": "test",
  "message": "Test key created! ...",
  "limits": {
    "max_file_size_mb": 10,
    "storage_mb": 100,
    "bandwidth_gb_per_month": 1,
    "rate_limit_per_hour": 100
  },
  "docs": "https://images.labnocturne.com/docs"
}
```

Present to the user:
- **API Key**: the `api_key` value
- **Type**: the `type` value
- **Limits**: list each limit from the `limits` object in human-readable form
- **How to use**: tell the user to set `export LABNOCTURNE_API_KEY=<api_key>` in their shell to use it with other skills

**Error** — the response body is:
```json
{
  "error": {
    "message": "Human-readable message",
    "type": "error_category",
    "code": "machine_readable_code"
  }
}
```

Show the error message to the user. Key generation rarely fails — if it does, suggest trying again in a moment.
