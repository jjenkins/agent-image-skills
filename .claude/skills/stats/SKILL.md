---
name: stats
description: "View storage usage and quota statistics for Lab Nocturne. Triggered by /stats or naturally (e.g. 'how much storage am I using')."
---

# Usage Statistics

View storage usage and quota statistics for your Lab Nocturne Images account.

## Invocation

`/stats` — also triggered naturally (e.g. "show my stats", "how much storage am I using", "check my usage").

## Instructions

Follow these steps exactly:

### 1. Resolve the API key

- Check if `$LABNOCTURNE_API_KEY` is set: run `echo $LABNOCTURNE_API_KEY`.
- If the variable is empty or unset, generate a test key automatically:
  ```
  curl -s https://images.labnocturne.com/key
  ```
  The response is JSON: `{"api_key": "ln_test_..."}`. Extract the `api_key` value and use it. Tell the user a temporary test key was generated.
- If the variable is set, use its value.

### 2. Resolve the base URL

- Use `$LABNOCTURNE_BASE_URL` if set, otherwise default to `https://images.labnocturne.com`.

### 3. Fetch statistics

Run:
```
curl -s -H "Authorization: Bearer <api_key>" <base_url>/stats
```

### 4. Handle the response

**Success** — the response body is:
```json
{
  "storage_used_bytes": 1234567,
  "storage_used_mb": 1.18,
  "file_count": 5,
  "quota_bytes": 104857600,
  "quota_mb": 100,
  "usage_percent": 1.18
}
```

Present to the user as a summary:
- **Files**: the `file_count` value
- **Storage used**: the `storage_used_mb` value formatted with units (e.g. "1.18 MB of 100 MB")
- **Usage**: the `usage_percent` value with a visual indicator (e.g. "1.18% used")
- If `usage_percent` >= 80, warn the user that they are approaching their storage limit
- If `usage_percent` >= 95, strongly warn the user that they are nearly out of storage and should delete old files or upgrade

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

Show the error message to the user and suggest a fix based on the error code:

| `code` | Suggested fix |
|---|---|
| `missing_api_key` | Set `$LABNOCTURNE_API_KEY` or let the skill generate a test key |
| `invalid_api_key` | Check that `$LABNOCTURNE_API_KEY` is correct, or unset it to auto-generate a test key |
| `invalid_auth_format` | The key should be passed as `Bearer <key>` — this is handled automatically |
