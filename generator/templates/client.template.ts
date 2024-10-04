type Scheme = "http" | "https";
type Hostname = `${number}.${number}.${number}.${number}` | string;

type ClientOpts = {
	host: `${Scheme}://${Hostname}`;
	port?: number;
	path?: string;
};

type ProcedureType = "query" | "mutation";
type Procedure = {
	payload: unknown;
	result: unknown;
};

type ProcedureResponse<Result = unknown> = {
	ok: boolean;
	error?: unknown;
	data?: Result;
};

type ProcedureSchema = Record<string, Procedure>;

type ClientSchema = { queries: ProcedureSchema; mutations: ProcedureSchema };

type SchemaBasedOnType<Schema extends ClientSchema, Type extends ProcedureType> = Schema[Type extends "query" ? "queries" : "mutations"];

type PayloadOf<Schema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<Schema, PType>> = SchemaBasedOnType<Schema, PType>[PName]["payload"];

type ResultOf<Schema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<Schema, PType>> = SchemaBasedOnType<Schema, PType>[PName]["result"];

type RawCallOpts<Schema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<Schema, PType>> = {
	name: PName;
	payload: PayloadOf<Schema, PType, PName>;
	extraHeaders?: Record<string, string>;
};

type CallOpts<Schema extends ClientSchema, PType extends ProcedureType, PName extends keyof SchemaBasedOnType<Schema, PType>> = Omit<
	Omit<RawCallOpts<Schema, "query", PName>, "name">,
	"payload"
>;

class ProcedureCallError extends Error {
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

class Client<Schema extends ClientSchema> {
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
	public static new<Schema extends ClientSchema>(opts: ClientOpts): Client<Schema> {
		return new Client<Schema>(opts);
	}

	// Get the client's endpoint
	public getEndpoint(): string {
		return this.endpoint;
	}

	// Manually call a robin procedure; this is a low-level function that should not be used directly unless absolutely necessary
	async call<PType extends ProcedureType, PName extends keyof SchemaBasedOnType<Schema, PType>>(
		type: PType,
		opts: RawCallOpts<Schema, PType, PName>
	): Promise<ResultOf<Schema, PType, PName>> {
		const url = this.makeRequestUrl(type, opts.name as string);
		const response = await fetch(url, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				...opts.extraHeaders,
			},
		});

		if (!response.ok) {
			throw new ProcedureCallError(`Failed to call procedure ${String(opts.name)} with status code ${response.status}`, String(opts.name));
		}

		const data = (await response.json()) as ProcedureResponse<ResultOf<Schema, PType, PName>>;
		if (!data.ok) {
			throw new ProcedureCallError(data.error, String(opts.name));
		}

		return data?.data as ResultOf<Schema, PType, PName>;
	}

	private makeRequestUrl(type: ProcedureType, name: string): string {
		//TODO: fix
		return `${this.endpoint}/${type}/${name}`;
	}

	// Manually call a robin query procedure
	async query<PName extends keyof SchemaBasedOnType<Schema, "query">>(
		name: PName,
		payload: PayloadOf<Schema, "query", PName>,
		opts?: CallOpts<Schema, "query", PName>
	): Promise<ResultOf<Schema, "query", PName>> {
		opts = opts || {};
		return this.call("query", { name, payload, ...opts });
	}

	// Manually call a robin mutation procedure
	async mutate<PName extends keyof SchemaBasedOnType<Schema, "mutation">>(
		name: PName,
		payload: PayloadOf<Schema, "mutation", PName>,
		opts?: CallOpts<Schema, "mutation", PName>
	): Promise<ResultOf<Schema, "mutation", PName>> {
		opts = opts || {};
		return this.call("mutation", { name, payload, ...opts });
	}
}

export type { ClientSchema, ProcedureSchema, ProcedureType, Procedure, ClientOpts, RawCallOpts, CallOpts, SchemaBasedOnType, PayloadOf, ResultOf };

export { ProcedureCallError };

export default Client;
