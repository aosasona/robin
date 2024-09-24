export type Schema = {
    queries: {
        list: {
            result: Array<string>;
            payload: void;
        };
    };
    mutations: {
        create: {
            result: string;
            payload: void;
        };
    };
};