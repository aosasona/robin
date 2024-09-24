type Scheme = "http" | "https";

type ClientOpts = {
  host: `${Scheme}://${string}`;
  port?: number;
  route: string;
};

class Client {
  private endpoint: string;

  public constructor(opts: ClientOpts) { }

  public static new(opts: ClientOpts): Client {
    return new Client(opts);
  }
}

export default Client;
