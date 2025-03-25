import { Env } from "./worker";

export class HelloDurable {
  constructor(private state: DurableObjectState, private env: Env) {
    // No special constructor logic needed for this example
  }

  async fetch(request: Request): Promise<Response> {
    // Handle CORS preflight requests
    if (request.method === "OPTIONS") {
      return new Response(null, {
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type",
        },
      });
    }

    // Return a simple string for any request
    return new Response("Hello, Texty.", {
      status: 200,
      headers: {
        "Content-Type": "text/plain",
        "Access-Control-Allow-Origin": "*",
      },
    });
  }
} 