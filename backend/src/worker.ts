import { TextDurable } from "./TextDurable";

export { TextDurable }; // Named export for Wrangler

export interface Env {
  TEXT_DO: DurableObjectNamespace;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    // Handle CORS preflight for the worker itself
    if (request.method === "OPTIONS") {
      return new Response(null, {
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type",
        },
      });
    }

    // Get a reference to our DO instance
    const id = env.TEXT_DO.idFromName("singleton");
    const stub = env.TEXT_DO.get(id);

    // Forward all requests to the DO
    // The DO will handle its own routing internally
    return stub.fetch(request);
  }
}; 