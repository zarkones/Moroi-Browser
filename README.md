# browser

A powerful, simple, and extendable browser library with three interfaces: use it as a Go library, drive it via CLI, or control it over HTTP via API.

Hard to detect via anti-bot mechanisms.
***

## What's in the box

| Interface | Use case |
|-----------|----------|
| **Library** | Embed in your Go agent/program directly |
| **CLI** | Interactive shell for manual browsing, debugging, or scripting |
| **API** | HTTP JSON server - call it from any language, any agent |

The browser instance stays alive across operations. Take a snapshot, get refs, click things, type text, screenshot - all in sequence without re-navigating.

***

Requires Chrome or Chromium to be installed and on your PATH.

***

## CLI

Start the interactive shell:

```bash
go run . --headless=false   # visible browser window
go run . --headless         # headless (no window)
```

### Commands

| Command | Description |
|---------|-------------|
| `navigate <url>` | Go to a URL |
| `snapshot` | Dump the AI-readable page tree, populates refs |
| `click <ref>` | Click an element by ref (e.g. `e3`) |
| `type <ref> <text...>` | Type into an input or textarea |
| `scroll <ref>` | Scroll element into viewport |
| `url` | Print current URL |
| `html` | Dump full page HTML |
| `text` | Dump visible page text |
| `back` | Navigate back |
| `forward` | Navigate forward |
| `refresh` | Reload the page |
| `screenshot <path>` | Save a PNG screenshot |
| `help` | Show command list |
| `quit` / `exit` / `close` | Close browser and exit |

### Workflow

Always run `snapshot` after navigating or after a page change. Refs (`e1`, `e42`, etc.) are only valid from the most recent snapshot.

```
browser> navigate https://github.com
browser> snapshot
browser> click e7
browser> snapshot
browser> type e12 hello world
browser> screenshot /tmp/result.png
```

***

## API

Start the HTTP server instead:

```bash
go run . --port 8080
go run . --headless --port 8080
```

All endpoints return JSON: `{"data": ...}` on success, `{"error": "..."}` on failure.

### Endpoints

| Method | Path | Body | Description |
|--------|------|------|-------------|
| `POST` | `/navigate` | `{"url":"https://..."}` | Navigate to URL |
| `GET` | `/snapshot` | - | Get page snapshot + update refs |
| `POST` | `/click` | `{"ref":"e3"}` | Click element by ref |
| `POST` | `/type` | `{"ref":"e5","text":"hello"}` | Type into element |
| `POST` | `/scroll` | `{"ref":"e9"}` | Scroll element into view |
| `GET` | `/url` | - | Get current URL |
| `GET` | `/html` | - | Get full page HTML |
| `GET` | `/text` | - | Get visible page text |
| `POST` | `/back` | - | Navigate back |
| `POST` | `/forward` | - | Navigate forward |
| `POST` | `/refresh` | - | Reload the page |
| `POST` | `/screenshot` | `{"path":"/tmp/shot.png"}` | Save screenshot to path |

### Example session

```bash
# Navigate
curl -X POST http://localhost:8080/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com"}'

# Snapshot (get refs)
curl http://localhost:8080/snapshot

# Click a ref from the snapshot
curl -X POST http://localhost:8080/click \
  -H "Content-Type: application/json" \
  -d '{"ref":"e7"}'

# Type into an input
curl -X POST http://localhost:8080/type \
  -H "Content-Type: application/json" \
  -d '{"ref":"e12","text":"chromedp"}'

# Screenshot
curl -X POST http://localhost:8080/screenshot \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/result.png"}'
```

***

## Library

Import and use directly in Go:

```go
b, err := browser.New(false)
if err != nil {
    log.Fatal(err)
}
defer b.Close()

b.Navigate("https://example.com")

snap, _ := b.Snapshot()
fmt.Println(snap)

b.ClickByRef("e3")
b.TypeByRef("e12", "hello world")
b.TakeScreenshot("/tmp/shot.png")
```

***

## How Snapshot works

`Snapshot()` runs a JS traversal of the live DOM and returns a hierarchical, role-annotated tree - the same format LLMs and agents parse naturally. Every interactive element gets a stable `ref` (e.g. `e3`, `e42`).

```
Snapshot:
- document:
  - navigation "Main nav" [ref=e1]:
    - link "Home" [ref=e2]:
      - /url: /
    - link "Docs" [ref=e3]:
      - /url: /docs
  - main:
    - heading "Get started" [ref=e4] [level=1]
    - textbox "" [ref=e5]:
      - /placeholder: Search...
    - button "Submit" [ref=e6]
```

Use refs to click, type, or scroll. Re-run `snapshot` after any navigation or DOM change to get fresh refs.

***

## Flags

```
--headless        Run without a visible browser window (default: false)
--port <n>        Port for API server mode (default: 8080)
```