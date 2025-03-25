# 1. High-Level Overview

1. **Goal**: A single shared “wall of text” that multiple users can edit in real-time via a simple web frontend (hosted on Cloudflare Pages).  
2. **Backend**: A Durable Object (DO) for real-time synchronization and concurrency control, backed by D1 for persistence.  
3. **Real-Time Updates**: SSE (Server-Sent Events) or WebSockets to push changes to all connected clients instantly. (Below, I’ll assume **SSE** for simplicity, since it’s straightforward to implement in a Worker context; WebSockets are also feasible but more complex to debug.)  
4. **Data Flow**:
   - Each user navigates to the Cloudflare Pages site, which loads the last known text (fetched from the DO or through a Worker proxy).  
   - The user’s browser then establishes an SSE connection to the DO.  
   - As a user types or otherwise modifies the text, the browser sends the diff (or entire text) to the DO via `fetch` request.  
   - The DO merges the change, persists it in memory, optionally flushes it to D1, and broadcasts the updated text to all connected clients via SSE.  
   - Every connected client updates its display in real time.

---

## 2. Project Structure

We’ll maintain a monorepo with two main parts: one for the **frontend** (Cloudflare Pages) and one for the **backend** (Durable Object + D1 + any additional Worker logic).

**Directory layout (example):**
my-cf-text-editor/ 
┣━ frontend/ // Deployed to Cloudflare Pages 
┣━ backend/ // Wrangler-based Worker project 
┗━ README.md


### 2.1. Frontend (Cloudflare Pages)

- **Build Tools**: A simple HTML + JS page or a minimal framework (e.g., React or Vue).  
- **Key Responsibilities**:
  1. **Render** the text (the “editor”).  
  2. **Initialize** an SSE connection to the DO to receive real-time updates.  
  3. **Handle** text input and send updates to the DO (via `fetch` POST).  
  4. **Dynamically update** the editor text upon receiving SSE messages.

### 2.2. Backend (Cloudflare Worker + Durable Object + D1)

- **Backend Worker** (optional layer):
  - Could act as a router/proxy to the Durable Object (not strictly necessary if the DO exposes a `fetch` directly).  
- **Durable Object**:
  - Holds the authoritative, in-memory version of the text.  
  - Handles concurrency by ensuring only one instance is active for this single “document.”  
  - Loads the text from D1 if needed on startup, merges changes, broadcasts updates, and writes back to D1.  
- **D1** (SQL Database):
  - Contains a table for storing the single text document, plus optional metadata.  
  - For a single shared doc, we might store:
    ```
    CREATE TABLE documents (
      id       TEXT PRIMARY KEY,
      content  TEXT NOT NULL
    );
    ```
  - The DO is responsible for read/write to this table.

---

## 3. Detailed Flow

### 3.1. Page Load
1. **Client** requests the page from **Cloudflare Pages**.  
2. **Pages** returns `index.html`, `main.js`, etc.

### 3.2. Initial Document Fetch
1. **Client** sends a GET request to the DO to fetch current text content.  
2. **DO** reads in-memory or from D1, returns the text.

### 3.3. Establish SSE Connection
1. Client opens an **SSE** connection: `new EventSource('/api/subscribe')`.  
2. **DO** keeps the connection open and stores the client in a set.  
3. DO sends SSE messages on updates, and client updates its local text.

### 3.4. Sending Edits
1. When the user edits text, the frontend sends a `POST` with either the whole text or a diff.  
2. **DO** merges changes, broadcasts the new text to all SSE clients, and writes to D1.

### 3.5. Persistence (Durable Object + D1)
- **Storing**:
  - DO writes updates to D1 table `documents` (row with `id='main'`).  
- **Loading**:
  - DO fetches `content` from the `documents` table on startup or first request.

### 3.6. Concurrency / Conflict Management
- One DO instance for the single document ensures concurrency is handled by Cloudflare’s single-threaded model for that DO.

---

## 4. Configuration & Deployment Artifacts

### 4.1. `wrangler.toml` (Backend)
Use a minimal example like:
\`\`\`toml
name = "my-do-text-editor"
compatibility_date = "2023-01-01"

[[d1_databases]]
binding = "DB"
database_name = "MY_D1_DB"
database_id = "<UUID for your DB>"

[[durable_objects]]
binding = "EDITOR_DO"
class_name = "EditorDurable"
script_name = "my-do-text-editor"

main = "dist/worker.js"
\`\`\`

### 4.2. `package.json` (Backend)
\`\`\`json
{
  "name": "my-do-text-editor",
  "scripts": {
    "start": "wrangler dev",
    "deploy": "wrangler publish"
  },
  "dependencies": {
    "@cloudflare/workers-types": "^4.20230518.0",
    "itty-router": "^3.0.12"
  },
  "devDependencies": {
    "typescript": "^4.9.3"
  }
}
\`\`\`

### 4.3. Durable Object Implementation
Example (TypeScript): **`src/EditorDurable.ts`**:
\`\`\`ts
export class EditorDurable {
  private clients: Set<ReadableStreamDefaultController> = new Set();
  private content: string = "";

  constructor(private state: DurableObjectState, private env: Env) {}

  async fetch(request: Request) {
    const url = new URL(request.url);

    if (!this.content) {
      const result = await this.env.DB.prepare(\`
        SELECT content FROM documents WHERE id = ?
      \`).bind('main').first();
      this.content = result?.content ?? "";
    }

    if (url.pathname === "/subscribe") {
      return this.handleSubscribe(request);
    } else if (url.pathname === "/edit" && request.method === "POST") {
      return this.handleEdit(request);
    } else if (url.pathname === "/content" && request.method === "GET") {
      return new Response(JSON.stringify({ content: this.content }), {
        headers: { "Content-Type": "application/json" },
      });
    }

    return new Response("Not Found", { status: 404 });
  }

  private async handleSubscribe(request: Request) {
    const { readable, writable } = new TransformStream();
    const writer = writable.getWriter();

    const headers = new Headers({
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    });

    const encoder = new TextEncoder();
    const clientController = {
      enqueue: (msg: string) => writer.write(encoder.encode(msg)),
      close: () => writer.close(),
    };
    this.clients.add(clientController);

    clientController.enqueue(\`event: update\\ndata: \${JSON.stringify({ content: this.content })}\\n\\n\`);

    request.signal.addEventListener("abort", () => {
      this.clients.delete(clientController);
      clientController.close();
    });

    return new Response(readable, { headers });
  }

  private async handleEdit(request: Request) {
    const body = await request.json();
    const newContent = body.content ?? "";

    this.content = newContent;
    this.broadcastUpdate(newContent);

    await this.env.DB.prepare(\`
      INSERT INTO documents (id, content)
      VALUES (?, ?)
      ON CONFLICT (id) DO UPDATE SET content=excluded.content
    \`).bind('main', newContent).run();

    return new Response("OK");
  }

  private broadcastUpdate(content: string) {
    const data = JSON.stringify({ content });
    for (const client of this.clients) {
      client.enqueue(\`event: update\\ndata: \${data}\\n\\n\`);
    }
  }
}
\`\`\`

### 4.4. Worker Entry Point
**`src/worker.ts`**:
\`\`\`ts
import { EditorDurable } from "./EditorDurable";

export interface Env {
  DB: D1Database;
  EDITOR_DO: DurableObjectNamespace;
}

export default {
  fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const id = env.EDITOR_DO.idFromName("singleton");
    const obj = env.EDITOR_DO.get(id);
    return obj.fetch(request);
  },

  DurableObject: EditorDurable,
};
\`\`\`

### 4.5. Cloudflare Pages (Frontend)
Minimal **`frontend/index.html`**:
\`\`\`html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Cloudflare Real-Time Editor</title>
</head>
<body>
  <textarea id="editor" style="width: 100%; height: 90vh;"></textarea>

  <script>
    const editor = document.getElementById('editor');
    let debounceTimer;

    // 1) Fetch initial content
    fetch('/content')
      .then(res => res.json())
      .then(data => {
        editor.value = data.content;
      });

    // 2) Set up SSE
    const eventSource = new EventSource('/subscribe');
    eventSource.addEventListener('update', (event) => {
      const data = JSON.parse(event.data);
      editor.value = data.content;
    });

    // 3) Listen for edits (debounced)
    editor.addEventListener('input', () => {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => {
        fetch('/edit', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({ content: editor.value })
        });
      }, 500);
    });
  </script>
</body>
</html>
\`\`\`

---

## 5. Best Practices & Considerations

1. **Performance**: SSE scales to many clients. Keep data minimal.  
2. **Security**: Consider authentication or other restrictions if needed.  
3. **Deployment**:  
   - Use `wrangler publish` for the Worker + DO.  
   - Deploy the Pages frontend from the `frontend` folder.  
4. **Persistence**: D1 is the permanent store; DO memory is ephemeral. Save frequently or on each edit.  
5. **Error Handling**: In production, handle DB errors or network issues gracefully.

---

## 6. Summary of Cloudflare Console Steps

1. **Create** a D1 database (e.g., `MY_D1_DB`) in your Cloudflare account.  
2. **Add** the `documents` table:

``` sql
CREATE TABLE documents ( id TEXT PRIMARY KEY, content TEXT NOT NULL ); INSERT INTO documents (id, content) VALUES ('main', '');
```

3. **Update** `wrangler.toml` with `database_id`.  
4. **Set** up a DO binding in the console if needed.  
5. **Configure** Cloudflare Pages for your `frontend/` folder.  
6. **Deploy** via `wrangler publish` (backend) and your Pages deployment flow (frontend).

---

## 7. Conclusion

This design leverages:

- **Cloudflare Pages** for static hosting,  
- **Durable Objects** for concurrency and real-time updates,  
- **D1** for persistent storage,  
- **SSE** for straightforward live updates.