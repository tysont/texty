interface Env {
  // Add any environment variables here if needed
}

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

  constructor(private state: DurableObjectState, private env: Env) {}

  async fetch(request: Request): Promise<Response> {
    // Handle CORS preflight requests
    if (request.method === "OPTIONS") {
      return new Response(null, {
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type",
        },
      });
    }

    const url = new URL(request.url);
    const path = url.pathname.replace(/^\/+/, "").replace(/\/+$/, "");
    
    console.log(`[DO] Handling ${request.method} request to path: "${path}" from URL: ${url.toString()}`);

    switch (path) {
      case "text": {
        if (request.method === "GET") {
          return this.handleGetText();
        } else if (request.method === "POST") {
          return this.handlePostText(request);
        }
        break;
      }
      case "subscribe": {
        if (request.method === "GET") {
          return this.handleSubscribe(request);
        }
        break;
      }
      case "lock/acquire": {
        if (request.method === "POST") {
          return this.handleAcquireLock(request);
        }
        break;
      }
      case "lock/release": {
        if (request.method === "POST") {
          return this.handleReleaseLock(request);
        }
        break;
      }
    }

    return new Response("Not Found", { status: 404 });
  }

  private handleGetText(): Response {
    return new Response(JSON.stringify({ 
      text: this.currentText,
      lockHolder: this.lockHolder 
    }), {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
      }
    });
  }

  private async handlePostText(request: Request): Promise<Response> {
    const body = await request.json() as RequestBody;
    const userId = body.userId;
    const text = body.text;

    // Only allow text updates from the lock holder
    if (!userId || userId !== this.lockHolder) {
      return new Response(JSON.stringify({ 
        success: false, 
        error: "You don't have the lock" 
      }), {
        status: 403,
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*"
        }
      });
    }

    this.currentText = text || "";
    this.broadcastState();

    return new Response(JSON.stringify({ success: true }), {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
      }
    });
  }

  private async handleAcquireLock(request: Request): Promise<Response> {
    const body = await request.json() as RequestBody;
    const userId = body.userId;

    if (!userId) {
      return new Response(JSON.stringify({ error: "No userId provided" }), {
        status: 400,
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*"
        }
      });
    }

    // If lock is free or belongs to same user, grant it
    if (!this.lockHolder || this.lockHolder === userId) {
      this.lockHolder = userId;
      this.broadcastState();
      return new Response(JSON.stringify({ success: true }), {
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*"
        }
      });
    }

    // Lock is taken by someone else
    return new Response(JSON.stringify({ 
      success: false, 
      error: "Lock is owned by another user" 
    }), {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
      }
    });
  }

  private async handleReleaseLock(request: Request): Promise<Response> {
    const body = await request.json() as RequestBody;
    const userId = body.userId;

    if (!userId) {
      return new Response(JSON.stringify({ error: "No userId provided" }), {
        status: 400,
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*"
        }
      });
    }

    // Only allow release if the user holds the lock
    if (this.lockHolder === userId) {
      this.lockHolder = "";
      this.broadcastState();
      return new Response(JSON.stringify({ success: true }), {
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*"
        }
      });
    }

    return new Response(JSON.stringify({ 
      success: false, 
      error: "Lock not owned by you" 
    }), {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
      }
    });
  }

  private handleSubscribe(request: Request): Response {
    console.log("[DO] Setting up SSE connection");
    
    const { readable, writable } = new TransformStream();
    const writer = writable.getWriter();
    const encoder = new TextEncoder();

    const headers = {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
      "Access-Control-Allow-Origin": "*"
    };

    const connectionController: SSEConnection = {
      enqueue(data: string) {
        writer.write(encoder.encode(data));
      },
      close() {
        writer.close();
      }
    };

    this.sseConnections.add(connectionController);
    console.log(`[DO] Added new SSE connection. Total connections: ${this.sseConnections.size}`);

    request.signal.addEventListener("abort", () => {
      this.sseConnections.delete(connectionController);
      connectionController.close();
      console.log(`[DO] Removed SSE connection. Remaining connections: ${this.sseConnections.size}`);
    });

    // Send initial state to new subscriber
    connectionController.enqueue(this.formatSSEData());
    console.log("[DO] Sent initial state to new SSE connection");

    return new Response(readable, { headers });
  }

  private broadcastState() {
    const message = this.formatSSEData();
    console.log(`[DO] Broadcasting update to ${this.sseConnections.size} connections`);
    for (const conn of this.sseConnections) {
      conn.enqueue(message);
    }
  }

  private formatSSEData(): string {
    const payload = JSON.stringify({
      text: this.currentText,
      lockHolder: this.lockHolder
    });
    return `event: update\ndata: ${payload}\n\n`;
  }
} 