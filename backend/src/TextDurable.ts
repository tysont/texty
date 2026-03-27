// ABOUTME: Durable Object that manages a single document's text and lock state.
// ABOUTME: Persists text to storage and broadcasts changes via SSE.

interface Env {}

interface RequestBody {
  text?: string;
  userId?: string;
}

interface SSEConnection {
  enqueue(data: string): void;
  close(): void;
}

export class TextDurable {
  private currentText: string = "";
  private lockHolder: string = "";
  private sseConnections: Set<SSEConnection> = new Set();
  private initialized: boolean = false;

  constructor(private state: DurableObjectState, private env: Env) {}

  private async ensureInitialized(): Promise<void> {
    if (this.initialized) return;
    this.currentText = (await this.state.storage.get<string>("text")) ?? "";
    this.lockHolder = (await this.state.storage.get<string>("lockHolder")) ?? "";
    this.initialized = true;
  }

  async fetch(request: Request): Promise<Response> {
    if (request.method === "OPTIONS") {
      return new Response(null, {
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type",
        },
      });
    }

    await this.ensureInitialized();

    const url = new URL(request.url);
    const path = url.pathname.replace(/^\/+/, "").replace(/\/+$/, "");

    switch (path) {
      case "text": {
        if (request.method === "GET") return this.handleGetText();
        if (request.method === "POST") return this.handlePostText(request);
        break;
      }
      case "subscribe": {
        if (request.method === "GET") return this.handleSubscribe(request);
        break;
      }
      case "lock/acquire": {
        if (request.method === "POST") return this.handleAcquireLock(request);
        break;
      }
      case "lock/release": {
        if (request.method === "POST") return this.handleReleaseLock(request);
        break;
      }
      case "summary": {
        if (request.method === "GET") return this.handleGetSummary();
        break;
      }
    }

    return new Response("Not Found", { status: 404 });
  }

  private handleGetText(): Response {
    return corsJSON({
      text: this.currentText,
      lockHolder: this.lockHolder,
    });
  }

  private handleGetSummary(): Response {
    const lineCount = this.currentText === "" ? 0 : this.currentText.split("\n").length;
    return corsJSON({
      lineCount,
      connectedUsers: this.sseConnections.size,
    });
  }

  private async handlePostText(request: Request): Promise<Response> {
    const body = (await request.json()) as RequestBody;
    const userId = body.userId;
    const text = body.text;

    if (!userId || userId !== this.lockHolder) {
      return corsJSON({ success: false, error: "You don't have the lock" }, 403);
    }

    this.currentText = text || "";
    await this.state.storage.put("text", this.currentText);
    this.broadcastState();

    return corsJSON({ success: true });
  }

  private async handleAcquireLock(request: Request): Promise<Response> {
    const body = (await request.json()) as RequestBody;
    const userId = body.userId;

    if (!userId) {
      return corsJSON({ error: "No userId provided" }, 400);
    }

    if (!this.lockHolder || this.lockHolder === userId) {
      this.lockHolder = userId;
      await this.state.storage.put("lockHolder", this.lockHolder);
      this.broadcastState();
      return corsJSON({ success: true });
    }

    return corsJSON({ success: false, error: "Lock is owned by another user" });
  }

  private async handleReleaseLock(request: Request): Promise<Response> {
    const body = (await request.json()) as RequestBody;
    const userId = body.userId;

    if (!userId) {
      return corsJSON({ error: "No userId provided" }, 400);
    }

    if (this.lockHolder === userId) {
      this.lockHolder = "";
      await this.state.storage.put("lockHolder", "");
      this.broadcastState();
      return corsJSON({ success: true });
    }

    return corsJSON({ success: false, error: "Lock not owned by you" });
  }

  private handleSubscribe(request: Request): Response {
    const { readable, writable } = new TransformStream();
    const writer = writable.getWriter();
    const encoder = new TextEncoder();

    const conn: SSEConnection = {
      enqueue(data: string) {
        writer.write(encoder.encode(data));
      },
      close() {
        writer.close();
      },
    };

    this.sseConnections.add(conn);

    request.signal.addEventListener("abort", () => {
      this.sseConnections.delete(conn);
      conn.close();
    });

    conn.enqueue(this.formatSSEData());

    return new Response(readable, {
      headers: {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache",
        "Connection": "keep-alive",
        "Access-Control-Allow-Origin": "*",
      },
    });
  }

  private broadcastState() {
    const message = this.formatSSEData();
    for (const conn of this.sseConnections) {
      conn.enqueue(message);
    }
  }

  private formatSSEData(): string {
    const payload = JSON.stringify({
      text: this.currentText,
      lockHolder: this.lockHolder,
    });
    return `event: update\ndata: ${payload}\n\n`;
  }
}

function corsJSON(data: unknown, status: number = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      "Content-Type": "application/json",
      "Access-Control-Allow-Origin": "*",
    },
  });
}
