// ABOUTME: Cloudflare Worker entry point that routes requests to Durable Objects.
// ABOUTME: Routes /docs to IndexDurable and /docs/:id/* to per-document TextDurables.

import { TextDurable } from "./TextDurable";
import { IndexDurable } from "./IndexDurable";
import { HelloDurable } from "./HelloDurable";

export { HelloDurable };
export { TextDurable };
export { IndexDurable };

export interface Env {
  TEXT_DO: DurableObjectNamespace;
  INDEX_DO: DurableObjectNamespace;
}

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, POST, DELETE, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type",
};

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    if (request.method === "OPTIONS") {
      return new Response(null, { headers: corsHeaders });
    }

    const url = new URL(request.url);
    const path = url.pathname.replace(/^\/+/, "").replace(/\/+$/, "");
    const segments = path.split("/");

    // Legacy routes: /text, /subscribe, /lock/* -> singleton TextDurable
    if (["text", "subscribe"].includes(segments[0]) || (segments[0] === "lock")) {
      const id = env.TEXT_DO.idFromName("singleton");
      const stub = env.TEXT_DO.get(id);
      return stub.fetch(request);
    }

    // /docs -> IndexDurable (list, create)
    if (path === "docs") {
      const id = env.INDEX_DO.idFromName("index");
      const stub = env.INDEX_DO.get(id);
      return stub.fetch(request);
    }

    // /docs/:docId -> DELETE to IndexDurable
    if (segments.length === 2 && segments[0] === "docs" && request.method === "DELETE") {
      const id = env.INDEX_DO.idFromName("index");
      const stub = env.INDEX_DO.get(id);
      return stub.fetch(request);
    }

    // /docs/:docId/* -> per-document TextDurable
    if (segments.length >= 3 && segments[0] === "docs") {
      const docId = segments[1];
      const subPath = segments.slice(2).join("/");

      const id = env.TEXT_DO.idFromName(docId);
      const stub = env.TEXT_DO.get(id);

      // Rewrite URL to strip /docs/:docId prefix
      const newUrl = new URL(request.url);
      newUrl.pathname = "/" + subPath;
      return stub.fetch(new Request(newUrl.toString(), request));
    }

    return new Response("Not Found", { status: 404 });
  },
};
