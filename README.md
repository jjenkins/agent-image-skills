# Lab Nocturne Images

Image storage for AI agents. Upload, retrieve, and manage images with a single curl call — no dashboards, no signup, no configuration.

As agents like Claude Code, NanoClaw, OpenClaw, and NanoBot become part of daily workflows, they need somewhere to put images. A user sends a screenshot to an agent on Discord; later they ask the agent to pull it up. A coding agent generates a chart and needs a shareable URL. A CI pipeline captures visual diffs and needs ephemeral hosting. Lab Nocturne is the image backend for these workflows — instant CDN URLs from a single API call.

## Why Agents Need Image Storage

- **Memory and recall** — A user sends a screenshot to NanoClaw on Discord and asks about it days later. Agents need persistent storage to hold visual context across conversations.
- **Sharing and collaboration** — An agent generates a chart, uploads it, and hands back a CDN URL. The user drops the link in Slack, a doc, or an email — no manual download/re-upload step.
- **Asset management** — Coding agents like Claude Code and OpenClaw working on web projects need a place to host images during development: logos, screenshots, mockups.
- **Transient artifacts** — Test keys give agents 7-day ephemeral storage. Perfect for one-off visualizations, debug screenshots, or CI artifacts that don't need to live forever.

### Agent-Native by Design

- **Claude Code** — Built-in `/upload`, `/files`, `/stats`, `/delete` skills
- **ChatGPT** — GPT Action with OpenAPI schema ([`integrations/chatgpt-action/`](integrations/chatgpt-action/))
- **Any agent** — One curl call is all it takes

### Setting Up Claude Code Skills

Install all five skills with the skills CLI:

```bash
npx skills add jjenkins/labnocturne-image-client
```

Or clone the repo and the skills are available automatically when Claude Code runs from the project directory:

```bash
git clone https://github.com/jjenkins/labnocturne-image-client.git
cd labnocturne-image-client
# /upload, /files, /stats, /delete, /generate-key are now available
```

Set `LABNOCTURNE_API_KEY` in your environment to use a specific key, or leave it unset and the skills will auto-generate a test key.

## Quick Install

### Python

```bash
pip install git+https://github.com/jjenkins/labnocturne-image-client.git#subdirectory=python
```

### Go

```bash
go get github.com/jjenkins/labnocturne-image-client/go/labnocturne
```

### Ruby

```ruby
# In Gemfile:
gem 'labnocturne', git: 'https://github.com/jjenkins/labnocturne-image-client', glob: 'ruby/*.gemspec'
```

### JavaScript/Node.js

```bash
git clone https://github.com/jjenkins/labnocturne-image-client.git
cd labnocturne-image-client/javascript
npm install && npm link
```

### PHP

```bash
composer require labnocturne/image-client
```

## Quick Start Examples

### Python

```python
from labnocturne import LabNocturneClient

# Generate test API key
api_key = LabNocturneClient.generate_test_key()

# Create client and upload
client = LabNocturneClient(api_key)
result = client.upload('photo.jpg')
print(f"Image URL: {result['url']}")
```

[Full Python Documentation →](python/README.md)

### Go

```go
import "github.com/jjenkins/labnocturne-image-client/go/labnocturne"

// Generate test API key
apiKey, _ := labnocturne.GenerateTestKey()

// Create client and upload
client := labnocturne.NewClient(apiKey)
result, _ := client.Upload("photo.jpg")
fmt.Println("Image URL:", result.URL)
```

[Full Go Documentation →](go/README.md)

### Ruby

```ruby
require 'labnocturne'

# Generate test API key
api_key = LabNocturne::Client.generate_test_key

# Create client and upload
client = LabNocturne::Client.new(api_key)
result = client.upload('photo.jpg')
puts "Image URL: #{result['url']}"
```

[Full Ruby Documentation →](ruby/README.md)

### JavaScript/Node.js

```javascript
import LabNocturneClient from 'labnocturne';

// Generate test API key
const apiKey = await LabNocturneClient.generateTestKey();

// Create client and upload
const client = new LabNocturneClient(apiKey);
const result = await client.upload('photo.jpg');
console.log('Image URL:', result.url);
```

[Full JavaScript Documentation →](javascript/README.md)

### PHP

```php
use LabNocturne\LabNocturneClient;

// Generate test API key
$apiKey = LabNocturneClient::generateTestKey();

// Create client and upload
$client = new LabNocturneClient($apiKey);
$result = $client->upload('photo.jpg');
echo "Image URL: {$result['url']}\n";
```

[Full PHP Documentation →](php/README.md)

## API Overview

All client libraries implement the same core methods:

| Method | Description |
|--------|-------------|
| `generateTestKey()` | Generate a test API key (static method) |
| `upload(filePath)` | Upload an image file |
| `listFiles(page, limit, sort)` | List uploaded files with pagination |
| `getStats()` | Get usage statistics |
| `deleteFile(imageId)` | Delete an image (soft delete) |

## How It Works

### 1. Get a Test API Key

No signup required for testing:

```bash
curl https://images.labnocturne.com/key
```

Returns a test key with:
- 10MB file size limit
- 7-day file retention
- Perfect for development

### 2. Upload an Image

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer ln_test_abc123..." \
  -F "file=@photo.jpg"
```

Returns a CDN URL you can use immediately.

### 3. Use Your Image

```html
<img src="https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg" alt="My image">
```

The URL is a CloudFront CDN URL - fast, global delivery.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/key` | Generate a test API key |
| POST | `/upload` | Upload an image file |
| GET | `/i/:id` | Retrieve an image (redirects to CDN) |
| GET | `/files` | List your uploaded files |
| GET | `/stats` | Get usage statistics |
| DELETE | `/i/:id` | Delete an image (soft delete) |

## Key Features

### Test Keys (`ln_test_*`)
- Generate instantly without signup
- 10MB file size limit
- Files expire after 7 days
- Perfect for development and testing

### Live Keys (`ln_live_*`)
- Require email + payment
- 100MB file size limit
- Files stored permanently
- Production-ready

## Supported Image Formats

- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- GIF (`.gif`)
- WebP (`.webp`)

## Language-Specific Documentation

Detailed documentation for each language:

- **[Python](python/README.md)** - Full Python client documentation with examples
- **[Go](go/README.md)** - Full Go client documentation with examples
- **[Ruby](ruby/README.md)** - Full Ruby client documentation with examples
- **[JavaScript/Node.js](javascript/README.md)** - Full JavaScript client documentation with examples
- **[PHP](php/README.md)** - Full PHP client documentation with examples
- **[curl](examples/curl/README.md)** - Command-line examples for testing

## Installation Methods

| Language | Method | Notes |
|----------|--------|-------|
| **Python** | `pip install git+...#subdirectory=python` | Direct GitHub install ✅ |
| **Go** | `go get github.com/.../go/labnocturne` | Native monorepo support ✅ |
| **Ruby** | Gemfile with `git:` + `glob:` | Bundler git support ✅ |
| **PHP** | `composer require labnocturne/image-client` | Packagist registry ✅ |
| **JavaScript** | Clone + `npm link` | npm subdirectory workaround |

## Error Handling

All APIs return JSON errors with helpful messages:

```json
{
  "error": {
    "message": "File size exceeds limit for test keys (10MB)",
    "type": "file_too_large",
    "code": "file_size_exceeded"
  }
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad request (invalid parameters)
- `401` - Unauthorized (invalid API key)
- `413` - File too large
- `500` - Server error

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See [LICENSE](LICENSE) for details.

## Links

- **API Documentation**: https://images.labnocturne.com/docs
- **GitHub Issues**: https://github.com/jjenkins/labnocturne-image-client/issues
- **Main Project**: https://github.com/jjenkins/labnocturne

## Need Help?

Check the language-specific README files for detailed examples and usage instructions.
