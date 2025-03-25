import { HelloDurable } from "./HelloDurable";

export interface Env {
  HELLO_DO: DurableObjectNamespace;
}

// Re-export the HelloDurable class so Wrangler can find it
export { HelloDurable };

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    // We identify the DO instance by a stable name, e.g. "singleton"
    const id = env.HELLO_DO.idFromName("singleton");
    const stub = env.HELLO_DO.get(id);

    // Forward the request to that Durable Object instance
    return stub.fetch(request);
  }
}; 