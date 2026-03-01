---
name: delete
description: "Soft-delete an image from Lab Nocturne. Triggered by /delete <image_id> or naturally (e.g. 'delete that image')."
---

# Delete Image

Soft-delete an image from your Lab Nocturne Images account.

## Invocation

`/delete <image_id>` — also triggered naturally (e.g. "delete image img_abc...", "remove that image").

## Instructions

Follow these steps exactly:

### 1. Resolve the image ID

- The user provides an image ID as the argument (e.g. `/delete img_01JXYZ...`).
- If no ID is given, ask the user which image to delete. Suggest using `/files` to find the ID.
- If the user provides a raw ULID without the `img_` prefix, prepend `img_` automatically.
- The ID should match the pattern `img_` followed by a ULID (26 alphanumeric characters).

### 2. Confirm with the user

- **This is a destructive action.** Before executing, tell the user which image ID will be deleted and ask for confirmation.
- Only proceed if the user confirms.

### 3. Resolve the API key

- Check if `$LABNOCTURNE_API_KEY` is set: run `echo $LABNOCTURNE_API_KEY`.
- If the variable is empty or unset, generate a test key automatically:
  ```
  curl -s https://images.labnocturne.com/key
  ```
  The response is JSON: `{"api_key": "ln_test_..."}`. Extract the `api_key` value and use it. Tell the user a temporary test key was generated.
- If the variable is set, use its value.

### 4. Resolve the base URL

- Use `$LABNOCTURNE_BASE_URL` if set, otherwise default to `https://images.labnocturne.com`.

### 5. Delete the image

Run:
```
curl -s -X DELETE -H "Authorization: Bearer <api_key>" <base_url>/i/<image_id>
```

### 6. Handle the response

**Success** — the response body is:
```json
{
  "success": true,
  "message": "File deleted successfully",
  "id": "img_..."
}
```

Present to the user:
- Confirm the image was deleted successfully
- Show the deleted image ID

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
| `file_not_found` | The image ID was not found. Use `/files` to list your images and find the correct ID |
