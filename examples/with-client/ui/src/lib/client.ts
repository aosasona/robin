import Client from "./bindings";

const client = Client.new({
	endpoint: "http://localhost:8081/_robin",
});

export default client;
