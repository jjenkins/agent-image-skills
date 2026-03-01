---
name: files
description: "List uploaded files from Lab Nocturne with pagination and sorting. Triggered by /files or naturally (e.g. 'show my images')."
---

# List Files

List uploaded files from your Lab Nocturne Images account with pagination and sorting.

## Invocation

`/files [options]` — also triggered naturally (e.g. "list my images", "show my files", "what have I uploaded").

Options can be specified naturally (e.g. `/files newest first`, `/files limit 10`, `/files page 2`).

## Instructions

Follow these steps exactly:

### 1. Parse options

Extract optional parameters from the user's input:
- **limit**: number of files to return (1-100, default 100)
- **offset**: number of files to skip for pagination (default 0)
- **sort**: one of `uploaded_at_desc` (default), `uploaded_at_asc`, `size_desc`, `size_asc`

Map natural language to sort values:
- "newest first", "most recent" → `uploaded_at_desc`
- "oldest first" → `uploaded_at_asc`
- "largest first", "biggest" → `size_desc`
- "smallest first" → `size_asc`

If the user asks for "page N", calculate offset as `(N - 1) * limit`.

### 2. Resolve the API key

- Check if `$LABNOCTURNE_API_KEY` is set: run `echo $LABNOCTURNE_API_KEY`.
- If the variable is empty or unset, generate a test key automatically:
  ```
  curl -s https://images.labnocturne.com/key
  ```
  The response is JSON: `{"api_key": "ln_test_..."}`. Extract the `api_key` value and use it. Tell the user a temporary test key was generated.
- If the variable is set, use its value.

### 3. Resolve the base URL

- Use `$LABNOCTURNE_BASE_URL` if set, otherwise default to `https://images.labnocturne.com`.

### 4. Fetch the file list

Run:
```
curl -s -H "Authorization: Bearer <api_key>" "<base_url>/files?limit=<limit>&offset=<offset>&sort=<sort>"
```

### 5. Handle the response

**Success** — the response body is:
```json
{
  "files": [
    {
      "id": "img_...",
      "url": "https://cdn.labnocturne.com/...",
      "filename": "photo.jpg",
      "size": 245678,
      "mime_type": "image/jpeg",
      "uploaded_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 50,
    "limit": 100,
    "offset": 0,
    "next": "https://..."
  }
}
```

Present to the user as a formatted list:
- For each file show: **ID**, **filename**, **size** (human-readable, e.g. "240 KB"), **type** (`mime_type`), **uploaded** (relative time, e.g. "2 days ago")
- After the list, show pagination info: "Showing X-Y of Z files"
- If there are more files (i.e. `pagination.next` is present), tell the user how to see the next page (e.g. "Use `/files offset <next_offset>` to see more")
- If the list is empty, tell the user they have no uploaded files

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
