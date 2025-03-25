# 1. High-Level Overview

1. **Goal**: A single shared "wall of text" that multiple users can edit in real-time via a simple web frontend (hosted on Cloudflare Pages).  
2. **Backend**: A Durable Object (DO) for real-time synchronization and concurrency control.  
3. **Real-Time Updates**: Server-Sent Events (SSE) to push changes to all connected clients instantly.  
4. **Data Flow**:
   - Each user navigates to the Cloudflare Pages site, which loads the last known text (fetched from the DO).  
   - The user's browser then establishes an SSE connection to the DO.  
   - As a user types, the browser sends the entire text to the DO via `fetch` request.  
   - The DO merges the change, persists it in memory, and broadcasts the updated text to all connected clients via SSE.  
   - Every connected client updates its display in real time.

---

## 2. Project Structure

We maintain a monorepo with two main parts: one for the **frontend** (Cloudflare Pages) and one for the **backend** (Durable Object + Worker logic).

**Directory layout:**
texty/
┣━ frontend/     // Deployed to Cloudflare Pages 
┣━ backend/      // Wrangler-based Worker project 
┣━ scripts/      // Utility scripts (e.g., kill-ports.sh)
┣━ docs/         // Documentation
┣━ package.json  // Root package.json for managing scripts
┗━ README.md

### 2.1. Frontend (Cloudflare Pages)

- **Build Tools**: Simple HTML + JS with minimal styling
- **Key Responsibilities**:
  1. **Render** the text editor with dark mode support
  2. **Initialize** an SSE connection to the DO to receive real-time updates
  3. **Handle** text input and send updates to the DO (via `fetch` POST)
  4. **Manage** collaborative locking system
  5. **Generate** unique user IDs per browser tab
  6. **Dynamically update** the editor text upon receiving SSE messages

### 2.2. Backend (Cloudflare Worker + Durable Object)

- **Backend Worker**:
  - Routes requests to the Durable Object
  - Handles CORS and request validation
- **Durable Object**:
  - Holds the authoritative, in-memory version of the text
  - Manages the collaborative locking system
  - Handles concurrency by ensuring only one instance is active
  - Broadcasts updates to all connected clients

---

## 3. Detailed Flow

### 3.1. Page Load
1. **Client** requests the page from **Cloudflare Pages**
2. **Pages** returns `index.html` with embedded styles and scripts

### 3.2. User ID Generation
1. **Client** generates or retrieves a system-wide ID from localStorage
2. **Client** generates a unique tab ID using `crypto.randomUUID()`
3. **Client** combines them into a unique `userId` (format: `systemUserId-tabId`)

### 3.3. Initial Document Fetch
1. **Client** sends a GET request to the DO to fetch current text content
2. **DO** returns the current in-memory text

### 3.4. Establish SSE Connection
1. Client opens an **SSE** connection: `new EventSource('/subscribe')`
2. **DO** keeps the connection open and stores the client in a set
3. DO sends SSE messages on updates, and client updates its local text

### 3.5. Collaborative Locking System
1. **Lock Acquisition**:
   - Triggered on first keystroke (not on focus)
   - Client sends POST to `/lock/acquire` with their `userId`
   - DO validates and grants lock if available
   - Client updates UI to show lock status

2. **Lock Release**:
   - Automatic after 3 seconds of inactivity
   - Manual on page unload
   - Final text save before release
   - Broadcasts lock release to all clients

3. **Lock State Management**:
   - DO tracks current lock holder
   - SSE broadcasts lock state changes
   - Clients update UI to show lock status
   - Disabled textarea when locked by another user

### 3.6. Text Updates
1. When user types:
   - Attempts to acquire lock if not held
   - Schedules text save after 500ms debounce
   - Schedules lock release after 3s inactivity
2. On save:
   - Sends POST to `/text` with current content
   - DO validates sender has lock
   - DO updates text and broadcasts to all clients

### 3.7. Persistence
- **In-Memory Storage**:
  - Text is stored in the Durable Object's memory
  - State persists as long as the DO instance is active
  - Note: Text will be reset if the DO instance is evicted

---

## 4. Configuration & Deployment Artifacts

### 4.1. `wrangler.toml` (Backend)
```toml
name = "texty-backend"
compatibility_date = "2024-01-01"

[[durable_objects]]
binding = "TEXT_DO"
class_name = "TextDurable"
script_name = "texty-backend"

main = "dist/worker.js"
```

### 4.2. `package.json` (Root)
```json
{
  "name": "texty",
  "scripts": {
    "dev-backend": "wrangler dev --config backend/wrangler.toml",
    "dev-frontend": "wrangler pages dev frontend",
    "dev-all": "concurrently \"npm run dev-backend\" \"npm run dev-frontend\"",
    "kill-all": "./scripts/kill-ports.sh \"$(node -p \"Object.values(require('./package.json').config).join(',')\")\"",
    "run-all": "npm run kill-all && npm run dev-all"
  },
  "config": {
    "backend-port": "8787",
    "frontend-port": "8788",
    "backend-debug-port": "9229",
    "frontend-debug-port": "9230"
  }
}
```

### 4.3. Durable Object Implementation
Key features in `TextDurable.ts`:
- Lock management with `/lock/acquire` and `/lock/release` endpoints
- Text updates with `/text` endpoint
- SSE subscription handling
- CORS support

### 4.4. Frontend Implementation
Key features in `index.html`:
- Dark mode styling
- Unique user ID generation per tab
- Lock-aware text editing
- Debounced text saving
- Automatic lock management
- SSE connection handling
- Cleanup on page unload

---

## 5. Best Practices & Considerations

1. **Performance**:
   - SSE scales to many clients
   - Debounced text saves (500ms)
   - Efficient lock management
   - Minimal data transfer

2. **User Experience**:
   - Dark mode support
   - Visual feedback for lock status
   - Smooth text updates
   - No lost keystrokes
   - Automatic lock release

3. **Error Handling**:
   - SSE reconnection logic
   - Lock acquisition failures
   - Network error handling
   - Cleanup on page unload

4. **Development**:
   - Local development setup
   - Port management
   - Debug port configuration
   - Concurrent server running

---

## 6. Summary of Cloudflare Console Steps

1. **Set** up a DO binding in the console
2. **Configure** Cloudflare Pages for the `frontend/` folder
3. **Deploy** via `wrangler publish` (backend) and Pages deployment (frontend)

---

## 7. Conclusion

This implementation provides:
- Real-time collaborative text editing
- Efficient lock-based concurrency control
- Clean user experience with dark mode
- Robust error handling and recovery
- Easy local development setup