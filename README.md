# Texty

A drop-dead simple collaborative text editor built to experiment with Cloudflare Pages and Durable Objects. Multiple users can edit the same text in real-time, with automatic locking to prevent conflicts.

## Features

- Real-time collaborative text editing
- Automatic locking system (first to type gets the lock)
- Dark mode UI
- Server-Sent Events for live updates
- No database required (in-memory storage)

## Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/texty.git
   cd texty
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development servers:
   ```bash
   npm run dev-all
   ```
   This will start:
   - Backend on http://localhost:8787
   - Frontend on http://localhost:8788

4. Open http://localhost:8788 in your browser

## Development Commands

- `npm run dev-all`: Start both frontend and backend servers
- `npm run dev-frontend`: Start only the frontend server
- `npm run dev-backend`: Start only the backend server
- `npm run kill-all`: Kill all development servers
- `npm run deploy-all`: Deploy all resources to Cloudflare

## How It Works

Texty uses:
- Cloudflare Pages for hosting the frontend
- Durable Objects for real-time synchronization and locking
- Server-Sent Events for live updates
- In-memory storage (text resets if the DO instance is evicted)

## License

MIT