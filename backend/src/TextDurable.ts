interface Env {
  // Add any environment variables here if needed
}

interface RequestBody {
  text: string;
}

interface SSEConnection {
  enqueue(data: string): void;
  close(): void;
}

export class TextDurable {
  private currentText: string = "";
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
    
    // Extract the path and remove leading/trailing slashes
    // This handles both /text and text paths
    const path = url.pathname.replace(/^\/+/, "").replace(/\/+$/, "");
    
    // Log the incoming request for debugging
    console.log(`[DO] Handling ${request.method} request to path: "${path}" from URL: ${url.toString()}`);

    // Handle root path by redirecting to frontend
    if (path === "") {
      const frontendPort = "8788"; // Local development port
      const frontendURL = url.hostname === "localhost" 
        ? `http://localhost:${frontendPort}`
        : "https://texty-frontend.pages.dev";
      return Response.redirect(frontendURL, 302);
    }

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
          console.log("[DO] Setting up SSE connection");
          return this.handleSubscribe(request);
        }
        break;
      }
    }

    console.log(`[DO] No handler found for path: "${path}"`);
    return new Response("Not Found", { status: 404 });
  }

  private handleGetText(): Response {
    return new Response(JSON.stringify({ text: this.currentText }), {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
      }
    });
  }

  private async handlePostText(request: Request): Promise<Response> {
    const body = await request.json() as RequestBody;
    this.currentText = body.text || "";

    // Broadcast to all connected clients
    this.broadcastUpdate(this.currentText);

    return new Response(JSON.stringify({ success: true }), {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
      }
    });
  }

  private handleSubscribe(request: Request): Response {
    console.log("[DO] Setting up SSE connection");
    
    // Create a TransformStream for SSE
    const { readable, writable } = new TransformStream();
    const writer = writable.getWriter();
    const encoder = new TextEncoder();

    // Set up SSE headers
    const headers = {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
      "Access-Control-Allow-Origin": "*"
    };

    // Create connection controller
    const connectionController: SSEConnection = {
      enqueue(data: string) {
        writer.write(encoder.encode(data));
      },
      close() {
        writer.close();
      }
    };

    // Add to active connections
    this.sseConnections.add(connectionController);
    console.log(`[DO] Added new SSE connection. Total connections: ${this.sseConnections.size}`);

    // Handle client disconnect
    request.signal.addEventListener("abort", () => {
      this.sseConnections.delete(connectionController);
      connectionController.close();
      console.log(`[DO] Removed SSE connection. Remaining connections: ${this.sseConnections.size}`);
    });

    // Send initial state to new subscriber
    connectionController.enqueue(`event: update\ndata: ${JSON.stringify({ text: this.currentText })}\n\n`);
    console.log("[DO] Sent initial state to new SSE connection");

    return new Response(readable, { headers });
  }

  private broadcastUpdate(newText: string) {
    const payload = JSON.stringify({ text: newText });
    const message = `event: update\ndata: ${payload}\n\n`;

    console.log(`[DO] Broadcasting update to ${this.sseConnections.size} connections`);
    for (const conn of this.sseConnections) {
      conn.enqueue(message);
    }
  }
} 