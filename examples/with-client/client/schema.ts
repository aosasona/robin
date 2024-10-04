export type Schema = {
	queries: {
		ping: {
			result: string;
			payload: void;
		};
		"todos.list": {
			result: Array<string>;
			payload: void;
		};
	};
	mutations: {
		"todo.create": {
			result: string;
			payload: void;
		};
	};
};
