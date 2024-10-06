export type RequestOpts = {
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH" | "OPTIONS" | "HEAD";
  headers?: Record<string, string>;
  body?: string;
}

export type HttpClientFn = (url: string, opts?: RequestOpts) => Promise<Response>;

export type ClientOpts = {
  // The full robin endpoint to connect to (e.g. http://localhost:8080/_robin)
  endpoint?: string;

  // Optional custom client function to use for making requests
  clientFn?: HttpClientFn;
};

export type ProcedureType = "query" | "mutation";

export type Procedure = {
  payload: unknown;
  result: unknown;
};

export type ServerResponse<Result = unknown> = {
  ok: boolean;
  error?: unknown;
  data?: Result;
};

export type ProcedureSchema = Record<string, Procedure>;

export type ClientSchema = { queries: ProcedureSchema; mutations: ProcedureSchema };

export type SchemaBasedOnType<CSchema extends ClientSchema, Type extends ProcedureType> = CSchema[Type extends "query" ? "queries" : "mutations"];

export type PayloadOf<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = SchemaBasedOnType<
  CSchema,
  PType
>[PName]["payload"];

export type ResultOf<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = SchemaBasedOnType<
  CSchema,
  PType
>[PName]["result"];

export type ProcedureResult<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = {
  ok: boolean;
  data?: ResultOf<CSchema, PType, PName>;
  error?: unknown;
}

export type RawCallOpts<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = {
  name: PName;
  payload: PayloadOf<CSchema, PType, PName>;
  extraHeaders?: Record<string, string>;
};

export type CallOpts<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = Omit<
  Omit<RawCallOpts<CSchema, "query", PName>, "name">,
  "payload"
>;


/** ================ GENERATED SCHEMA ================ **/
export type Schema = {
    queries: {
        "whoami": {
            result: {
                user_id: string;
                created_at: string;
            };
            payload: void;
        };
        "list-todos": {
            result: Array<string>;
            payload: void;
        };
    };
    mutations: {
        "sign-in": {
            result: {
                user_id: string;
                created_at: string;
            } | null;
            payload: void;
        };
        "sign-up": {
            result: {
                user_id: string;
                created_at: string;
            } | null;
            payload: void;
        };
        "create-todo": {
            result: {
                title: string;
            };
            payload: {
                title: string;
            };
        };
    };
};


// Default client function that uses the fetch API which is available in most environments
export function defaultClientFn(url: string, opts?: RequestOpts): Promise<Response> {
  return fetch(url, {
    method: opts?.method || "GET",
    headers: opts?.headers || {},
    body: opts?.body || undefined,
  });
}

/**
 * ==================== CONTAINERS ====================
 *
 * These classes are used to group query and mutation methods together
 */
class Queries<CSchema extends ClientSchema = Schema> {
  constructor(private client: Client<CSchema>) {}
  
  /**
   * @procedure whoami
   *
   * @returns Promise<ProcedureResult<CSchema, "query", "whoami">>
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async whoami(opts?: CallOpts<CSchema, "query", "whoami">): Promise<ProcedureResult<CSchema, "query", "whoami">> {
    return await this.client.call("query", { ...opts, name: "whoami", payload: undefined });
  }

  /**
   * @procedure list-todos
   *
   * @returns Promise<ProcedureResult<CSchema, "query", "list-todos">>
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async listTodos(opts?: CallOpts<CSchema, "query", "list-todos">): Promise<ProcedureResult<CSchema, "query", "list-todos">> {
    return await this.client.call("query", { ...opts, name: "list-todos", payload: undefined });
  }
}

class Mutations<CSchema extends ClientSchema = Schema> {
  constructor(private client: Client<CSchema>) {}
  
  /**
   * @procedure sign-in
   *
   * @returns Promise<ProcedureResult<CSchema, "query", "sign-in">>
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async signIn(opts?: CallOpts<CSchema, "mutation", "sign-in">): Promise<ProcedureResult<CSchema, "mutation", "sign-in">> {
    return await this.client.call("mutation", { ...opts, name: "sign-in", payload: undefined });
  }

  /**
   * @procedure sign-up
   *
   * @returns Promise<ProcedureResult<CSchema, "query", "sign-up">>
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async signUp(opts?: CallOpts<CSchema, "mutation", "sign-up">): Promise<ProcedureResult<CSchema, "mutation", "sign-up">> {
    return await this.client.call("mutation", { ...opts, name: "sign-up", payload: undefined });
  }

  /**
   * @procedure create-todo
   *
   * @returns Promise<ProcedureResult<CSchema, "query", "create-todo">>
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async createTodo(payload: PayloadOf<CSchema, "mutation", "create-todo">, opts?: CallOpts<CSchema, "mutation", "create-todo">): Promise<ProcedureResult<CSchema, "mutation", "create-todo">> {
    return await this.client.call("mutation", { ...opts, name: "create-todo", payload: payload });
  }
}

/** ==================== CLIENT ==================== **/
class Client<CSchema extends ClientSchema = Schema> {
  private endpoint: string;
  private clientFn: HttpClientFn;

  public readonly queries: Queries<CSchema>;
  public readonly mutations: Mutations<CSchema>;

  public constructor(opts: ClientOpts) {
    if (!opts.endpoint) {
      throw new Error("An endpoint is required to create a new client");
    }

    this.endpoint = opts.endpoint;
    this.clientFn = opts.clientFn || defaultClientFn;

    this.queries = new Queries<CSchema>(this);
    this.mutations = new Mutations<CSchema>(this);
  }

  // Create a new client instance
  // deno-lint-ignore no-misused-new
  public static new<CSchema extends ClientSchema = Schema>(opts: ClientOpts): Client<CSchema> {
    return new Client<CSchema>(opts);
  }

  // Get the client's endpoint
  public getEndpoint(): string {
    return this.endpoint;
  }

  /**
   * @param {PType} type The type of the procedure to call
   * @param {RawCallOpts<CSchema, PType, PName>} opts The options for the procedure call
   * @returns Promise<ProcedureResult<CSchema, PType, PName>>
   *
   * @description Manually call a robin procedure; this is a low-level function that should not be used directly unless absolutely necessary
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async call<PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>>(
    type: PType,
    opts: RawCallOpts<CSchema, PType, PName>
  ): Promise<ProcedureResult<CSchema, PType, PName>> {
    try {
      const url = this.makeRequestUrl(type, String(opts.name));

      const requestOpts: RequestOpts = {
        method: "POST",
        body: opts.payload ? JSON.stringify({d: opts.payload}) : undefined,
        headers: {
          "Content-Type": "application/json",
          ...opts.extraHeaders,
        },
      };

      const response = await this.clientFn(url, requestOpts);
      if (!response.ok) {
        throw new ProcedureCallError(`Failed to call procedure \`${String(opts.name)}\` with status code ${response.status}`, String(opts.name));
      }

      const data = (await response.json()) as ServerResponse<ResultOf<CSchema, PType, PName>>;
      if (!data.ok) {
        return { ok: false, error: data?.error || "An unknown error occurred" };
      }

      return { ok: true, data: data?.data as ResultOf<CSchema, PType, PName> };
    } catch (e) {
      if (e instanceof ProcedureCallError) {
        throw e;
      }

      const message = Object.prototype.hasOwnProperty.call(e, "message") ? e.message : "An unknown error occurred";
      throw new ProcedureCallError(message, String(opts.name), e);
    }
  }

  /**
   * @param {PName} name The name of the query procedure to call
   * @param {PayloadOf<CSchema, "query", PName>} payload The payload to send to the query procedure
   * @param {CallOpts<CSchema, "query", PName>} opts The options for the query procedure call
   * @returns Promise<ProcedureResult<CSchema, "query", PName>>
   *
   * @description Manually call a robin query procedure
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async query<PName extends keyof SchemaBasedOnType<CSchema, "query">>(
    name: PName,
    payload: PayloadOf<CSchema, "query", PName>,
    opts?: CallOpts<CSchema, "query", PName>
  ): Promise<ProcedureResult<CSchema, "query", PName>> {
    opts = opts || {};
    return await this.call("query", { name, payload, ...opts });
  }

  /**
   * @param {PName} name The name of the mutation procedure to call
   * @param {PayloadOf<CSchema, "mutation", PName>} payload The payload to send to the mutation procedure
   * @param {CallOpts<CSchema, "mutation", PName>} opts The options for the mutation procedure call
   * @returns Promise<ProcedureResult<CSchema, "mutation", PName>>
   *
   * @description Manually call a robin mutation procedure
   * @throws {ProcedureCallError} if the procedure call fails
   */
  async mutate<PName extends keyof SchemaBasedOnType<CSchema, "mutation">>(
    name: PName,
    payload: PayloadOf<CSchema, "mutation", PName>,
    opts?: CallOpts<CSchema, "mutation", PName>
  ): Promise<ProcedureResult<CSchema, "mutation", PName>> {
    opts = opts || {};
    return await this.call("mutation", { name, payload, ...opts });
  }

  private makeRequestUrl(type: ProcedureType, name: string): string {
    const procType = type === "query" ? "q" : "m";
    return `${this.endpoint}?__proc=${procType}__${name}`;
  }
}

// Custom error class for procedure call errors
export class ProcedureCallError extends Error {
  // The actual error message from the server - in most cases, this will be a string, but it can be anything
  public details: unknown;

  // The name of the procedure that caused this error
  public procedureName: string;

  // The previous error that caused this error, if any
  public previousError: Error | null;

  public constructor(message: any, procedureName: string, originalError: Error | null = null) {
    super(typeof message === "string" ? message : "A procedure call error occurred, see the `details` property for more information");
    this.name = "ProcedureCallError";
    this.details = message;
    this.procedureName = procedureName;
    this.previousError = originalError;
  }

  public toString(): string {
    return `${this.name}: ${this.message}`;
  }
}

export default Client;
