import Client from "./bindings.ts";

const client = Client.new({
	endpoint: "http://localhost:8060",
});

await client.queries.ping();

const todos = await client.queries.todosList();
const newTodo = await client.mutations.todosCreate({
	title: "Buy milk",
	task_description: "Buy milk from the store",
	completed: false,
});

console.log("todos -> ", todos);
console.log("newTodo -> ", newTodo);

const raw = await client.mutations.raw({
	title: "Buy milk",
	task_description: "Buy milk from the store",
	completed: false,
});
const expected =
	'{"d":{"title":"Buy milk","task_description":"Buy milk from the store","completed":false}}';
if (raw !== expected) {
	console.error("raw mutation failed");
} else {
	console.log("raw mutation succeeded");
}

// This should throw since the generated client is set to throw on errors
// await client.queries.fail();
