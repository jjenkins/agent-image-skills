---
name: upload
description: "Upload an image to Lab Nocturne and get a CDN URL. Triggered by /upload <path> or naturally (e.g. 'upload this screenshot')."
---

# Upload Image to Lab Nocturne

Upload an image file to the Lab Nocturne Images API and return the CDN URL.

## Invocation

`/upload <path>` — where `<path>` is the file to upload. Also triggered naturally (e.g. "upload this screenshot").

## Instructions

Follow these steps exactly:

### 1. Resolve the file path

- The user provides a file path as the argument (e.g. `/upload ./screenshot.png`).
- If no path is given, ask the user which file to upload.
- Verify the file exists using `ls` on the path. If it does not exist, tell the user and stop.
- Verify the extension is one of: `jpg`, `jpeg`, `png`, `gif`, `webp`, `svg`. If not, tell the user the file type is not supported and list the allowed types.

### 2. Resolve the API key

- Check if `$LABNOCTURNE_API_KEY` is set: run `echo $LABNOCTURNE_API_KEY`.
- If the variable is empty or unset, generate a test key automatically:
  ```
  curl -s https://images.labnocturne.com/key
  ```
  The response is JSON: `{"api_key": "ln_test_..."}`. Extract the `api_key` value and use it for the upload. Tell the user a temporary test key was generated (7-day file retention, 10MB limit).
- If the variable is set, use its value.

### 3. Resolve the base URL

- Use `$LABNOCTURNE_BASE_URL` if set, otherwise default to `https://images.labnocturne.com`.

### 4. Upload the file

Run:
```
curl -s -X POST \
  -F "file=@<resolved_path>" \
  -H "Authorization: Bearer <api_key>" \
  <base_url>/upload
```

### 5. Handle the response

**Success** (HTTP 201) — the response body is:
```json
{
  "id": "img_...",
  "url": "https://cdn.labnocturne.com/...",
  "size": 123456,
  "mime_type": "image/png",
  "uploaded_at": "2025-01-01T00:00:00Z"
}
```

Present to the user:
- **Image ID**: the `id` field
- **CDN URL**: the `url` field
- **Size**: the `size` field, formatted as human-readable (e.g. "1.2 MB")

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
| `missing_file` | Internal skill error — report to user |
| `file_size_exceeded` | Test keys are limited to 10MB, live keys to 100MB. Compress the image or upgrade |
| `quota_exceeded` | Storage quota full. Delete old files with `curl -X DELETE <base_url>/i/<id>` or upgrade |
| `unsupported_file_type` | Only jpg, png, gif, webp, svg are supported |
| `upload_failed` | Server error — try again in a moment |
