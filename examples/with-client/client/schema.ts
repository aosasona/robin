export type Schema = {
    queries: {
        "todos.list": {
            result: Array<string>;
            payload: void;
        };
        "ping": {
            result: string;
            payload: string;
        };
    };
    mutations: {
        "todos.create": {
            result: string;
            payload: void;
        };
    };
};