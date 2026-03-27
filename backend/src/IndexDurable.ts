// ABOUTME: Durable Object that maintains an index of all documents.
// ABOUTME: Stores document metadata and provides list/create/delete operations.

interface DocEntry {
  name: string;
  createdAt: string;
}

export class IndexDurable {
  constructor(private state: DurableObjectState) {}

  async fetch(request: Request): Promise<Response> {
    if (request.method === "OPTIONS") {
      return corsResponse(null);
    }

    const url = new URL(request.url);
    const path = url.pathname.replace(/^\/+/, "").replace(/\/+$/, "");

    switch (request.method) {
      case "GET": {
        if (path === "docs") {
          return this.handleList();
        }
        break;
      }
      case "POST": {
        if (path === "docs") {
          return this.handleCreate(request);
        }
        break;
      }
      case "DELETE": {
        const match = path.match(/^docs\/([^/]+)$/);
        if (match) {
          return this.handleDelete(match[1]);
        }
        break;
      }
    }

    return new Response("Not Found", { status: 404 });
  }

  private async handleList(): Promise<Response> {
    const entries = await this.state.storage.list<DocEntry>({ prefix: "doc:" });
    const docs: { id: string; name: string; createdAt: string }[] = [];

    for (const [key, entry] of entries) {
      const id = key.replace("doc:", "");
      docs.push({ id, name: entry.name, createdAt: entry.createdAt });
    }

    return corsResponse(JSON.stringify({ docs }));
  }

  private async handleCreate(request: Request): Promise<Response> {
    const body = (await request.json()) as { name?: string };
    const name = body.name?.trim();

    if (!name) {
      return corsResponse(JSON.stringify({ error: "Name is required" }), 400);
    }

    // Slugify: lowercase, replace non-alphanumeric with hyphens
    const id = name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "");

    if (!id) {
      return corsResponse(JSON.stringify({ error: "Invalid name" }), 400);
    }

    const existing = await this.state.storage.get<DocEntry>(`doc:${id}`);
    if (existing) {
      return corsResponse(JSON.stringify({ error: "Document already exists" }), 409);
    }

    const entry: DocEntry = { name, createdAt: new Date().toISOString() };
    await this.state.storage.put(`doc:${id}`, entry);

    return corsResponse(JSON.stringify({ id, name }), 201);
  }

  private async handleDelete(id: string): Promise<Response> {
    const existing = await this.state.storage.get<DocEntry>(`doc:${id}`);
    if (!existing) {
      return corsResponse(JSON.stringify({ error: "Not found" }), 404);
    }

    await this.state.storage.delete(`doc:${id}`);
    return corsResponse(JSON.stringify({ success: true }));
  }
}

function corsResponse(body: string | null, status: number = 200): Response {
  return new Response(body, {
    status,
    headers: {
      "Content-Type": "application/json",
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, DELETE, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
    },
  });
}
