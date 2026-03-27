# Texty

A collaborative terminal text editor built with Go (Bubble Tea) and Cloudflare Workers (Durable Objects). Multiple users can edit shared documents in real-time from their terminals, with automatic locking to prevent conflicts.

## Features

- Terminal UI with dark theme and purple accents
- Multi-document support (create, browse, delete)
- Real-time collaborative editing via Server-Sent Events
- Pessimistic locking (first to type gets the lock, auto-releases after 3s idle)
- Presence indicators (see who's connected)
- Persistent storage (text survives server restarts)
- Server-side lock timeout (safety net for crashed clients)
- Username prompt on first launch

## Quick Start

1. Install the TUI client (requires Go 1.24+):
   ```bash
   cd tui
   go install ./cmd/texty
   ```

2. Start the backend:
   ```bash
   npm install
   npm run dev-backend
   ```

3. Run the editor:
   ```bash
   texty
   ```

   To connect to a different server:
   ```bash
   texty --server https://texty-backend.example.com
   ```

## Keyboard Shortcuts

### Document List
| Key | Action |
|-----|--------|
| `j` / `down` | Move cursor down |
| `k` / `up` | Move cursor up |
| `enter` | Open document |
| `n` | Create new document |
| `d` | Delete document |
| `r` | Refresh list |
| `q` | Quit |

### Editor
| Key | Action |
|-----|--------|
| Arrow keys | Move cursor |
| `Home` / `End` | Start / end of line |
| `Ctrl+A` / `Ctrl+E` | Start / end of line |
| `Enter` | New line |
| `Backspace` | Delete before cursor |
| `Delete` | Delete at cursor |
| `Tab` | Insert spaces |
| `Ctrl+S` | Force save |
| `Ctrl+Q` | Back to document list |
| `?` | Toggle help overlay |

## Development

### Commands

- `npm run dev-backend`: Start the backend on http://localhost:8787
- `npm run dev-frontend`: Start the legacy web frontend on http://localhost:8788
- `npm run dev-all`: Start both servers
- `npm run kill-all`: Kill all development servers
- `npm run deploy-all`: Deploy all resources to Cloudflare

### Architecture

- **TUI client** (`tui/`): Go + Bubble Tea + Lip Gloss. Communicates with the backend via REST and SSE.
- **Backend** (`backend/`): Cloudflare Workers + Durable Objects (TypeScript).
  - `TextDurable`: Per-document DO handling text, locks, SSE, and persistence.
  - `IndexDurable`: Singleton DO maintaining the document index.
  - Worker routes `/docs/:id/*` to per-document DOs and `/docs` to the index.
- **Web frontend** (`frontend/`): Legacy single-page HTML app (still functional).

### API

All document endpoints are scoped under `/docs/:docId/`:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/docs` | List all documents |
| `POST` | `/docs` | Create document |
| `DELETE` | `/docs/:id` | Delete document |
| `GET` | `/docs/:id/text` | Get text and lock state |
| `POST` | `/docs/:id/text` | Update text |
| `POST` | `/docs/:id/lock/acquire` | Acquire editing lock |
| `POST` | `/docs/:id/lock/release` | Release editing lock |
| `GET` | `/docs/:id/subscribe` | SSE stream (with `userId` and `username` query params) |

## License

MIT
