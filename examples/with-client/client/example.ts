import Client from "./bindings.ts";

const client = Client.new({
	endpoint: "http://localhost:8081/_robin",
});

const { data: pong } = await client.queries.ping("hello");
console.log("pong -> ", pong);

const { data: todos } = await client.queries.todosList();
console.log("todos -> ", todos);

const { data: newTodo } = await client.mutations.todosCreate({
	title: "Buy milk",
});
console.log("newTodo -> ", newTodo);
