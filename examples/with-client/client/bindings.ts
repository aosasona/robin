type Scheme = "http" | "https";
type Hostname = `${number}.${number}.${number}.${number}` | string;

export type ClientOpts = {
  host: `${Scheme}://${Hostname}`;
  port?: number;
  path?: string;
};

export type ProcedureType = "query" | "mutation";

export type Procedure = {
  payload: unknown;
  result: unknown;
};

export type ProcedureResponse<Result = unknown> = {
  ok: boolean;
  error?: unknown;
  data?: Result;
};

export type ProcedureSchema = Record<string, Procedure>;

export type ClientSchema = { queries: ProcedureSchema; mutations: ProcedureSchema };

export type SchemaBasedOnType<CSchema extends ClientSchema, Type extends ProcedureType> = CSchema[Type extends "query" ? "queries" : "mutations"];

export type PayloadOf<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = SchemaBasedOnType<CSchema, PType>[PName]["payload"];

export type ResultOf<CSchema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>> = SchemaBasedOnType<CSchema, PType>[PName]["result"];

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
        "ping": {
            result: string;
            payload: string;
        };
        "todos.list": {
            result: Array<string>;
            payload: void;
        };
    };
    mutations: {
        "todos.create": {
            result: string;
            payload: void;
        };
    };
};


// The main client class that will be used to interact with the robin procedures
class Client<CSchema extends ClientSchema = Schema> {
  private endpoint: string;

  public constructor(opts: ClientOpts) {
    let endpoint: string;
    endpoint = `${opts.host}`;

    if (opts.port !== undefined && opts.port > 0 && opts.port < 65536 && ![80, 443].includes(opts.port)) {
      endpoint += `:${opts.port}`;
    }

    if (opts.path != undefined && !opts.path.startsWith("/") && opts.path.length > 0) {
      endpoint += `/${opts.path}`;
    }

    this.endpoint = endpoint;
  }

  // Create a new client instance
  public static new<CSchema extends ClientSchema = Schema>(opts: ClientOpts): Client<CSchema> {
    return new Client<CSchema>(opts);
  }

  // Get the client's endpoint
  public getEndpoint(): string {
    return this.endpoint;
  }

  // Manually call a robin procedure; this is a low-level function that should not be used directly unless absolutely necessary
  async call<PType extends ProcedureType, PName extends keyof SchemaBasedOnType<CSchema, PType>>(
    type: PType,
    opts: RawCallOpts<CSchema, PType, PName>
  ): Promise<ResultOf<CSchema, PType, PName>> {
    const url = this.makeRequestUrl(type, opts.name as string);

    const fetchOpts: RequestInit = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...opts.extraHeaders,
      },
    };

    if (opts.payload !== undefined) {
      fetchOpts.body = JSON.stringify(opts.payload);
    }

    const response = await fetch(url, fetchOpts);
    if (!response.ok) {
      throw new ProcedureCallError(`Failed to call procedure ${String(opts.name)} with status code ${response.status}`, String(opts.name));
    }

    const data = (await response.json()) as ProcedureResponse<ResultOf<CSchema, PType, PName>>;
    if (!data.ok) {
      throw new ProcedureCallError(data.error, String(opts.name));
    }

    return data?.data as ResultOf<CSchema, PType, PName>;
  }

  // Manually call a robin query procedure
  async query<PName extends keyof SchemaBasedOnType<CSchema, "query">>(
    name: PName,
    payload: PayloadOf<CSchema, "query", PName>,
    opts?: CallOpts<CSchema, "query", PName>
  ): Promise<ResultOf<CSchema, "query", PName>> {
    opts = opts || {};
    return await this.call("query", { name, payload, ...opts });
  }

  // Manually call a robin mutation procedure
  async mutate<PName extends keyof SchemaBasedOnType<CSchema, "mutation">>(
    name: PName,
    payload: PayloadOf<CSchema, "mutation", PName>,
    opts?: CallOpts<CSchema, "mutation", PName>
  ): Promise<ResultOf<CSchema, "mutation", PName>> {
    opts = opts || {};
    return await this.call("mutation", { name, payload, ...opts });
  }

  private makeRequestUrl(type: ProcedureType, name: string): string {
    let procType = type === "query" ? "q" : "mutation";
    return `${this.endpoint}?__proc=${procType}__${name}`;
  }

  /** ================ GENERATED METHODS ================ **/

  
  /** @procedure "ping" */
  async ping(payload: PayloadOf<CSchema, "query", "ping">, opts?: CallOpts<CSchema, "query", "ping">): Promise<ResultOf<CSchema, "query", "ping">> {
    return await this.call("query", { name: "ping", payload: payload, ...opts});
  }

  /** @procedure "todos.list" */
  async todosList(opts?: CallOpts<CSchema, "query", "todos.list">): Promise<ResultOf<CSchema, "query", "todos.list">> {
    return await this.call("query", { name: "todos.list", payload: undefined, ...opts});
  }

  /** @procedure "todos.create" */
  async todosCreate(opts?: CallOpts<CSchema, "mutation", "todos.create">): Promise<ResultOf<CSchema, "mutation", "todos.create">> {
    return await this.call("mutation", { name: "todos.create", payload: undefined, ...opts});
  }
}

// Custom error class for procedure call errors
export class ProcedureCallError extends Error {
  // The actual error message from the server - in most cases, this will be a string, but it can be anything
  public details: unknown;

  // The name of the procedure that caused this error
  public procedureName: string;

  public constructor(message: any, procedureName: string) {
    super(typeof message === "string" ? message : "A procedure call error occurred, see the `details` property for more information");
    this.name = "RobinError";
    this.details = message;
    this.procedureName = procedureName;
  }

  public toString(): string {
    return `${this.name}: ${this.message}`;
  }
}

export default Client;
