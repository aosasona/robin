export type Schema = {
    queries: {
        ping: {
            result: string;
            payload: void;
        };
        getUser: {
            result: {
                id?: number;
                name: string;
            };
            payload: number;
        };
        getUsersByIds: {
            result: Array<{
                id?: number;
                name: string;
            }>;
            payload: Array<number>;
        };
        getUsers: {
            result: Array<{
                id?: number;
                name: string;
            }>;
            payload: void;
        };
    };
    mutations: {
        addUser: {
            result: {
                id?: number;
                name: string;
            };
            payload: {
                id?: number;
                name: string;
            };
        };
        deleteUser: {
            result: {
                id?: number;
                name: string;
            };
            payload: number;
        };
    };
};