import Client from "./bindings.ts";

const client = Client.new({
	endpoint: "http://localhost:8081/_robin",
});

await client.queries.ping();

const { data: todos } = await client.queries.todosList();
const { data: newTodo } = await client.mutations.todosCreate({
	title: "Buy milk",
	completed: false,
});

console.log("todos -> ", todos);
console.log("newTodo -> ", newTodo);
