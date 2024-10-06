import { createResource, Show } from "solid-js";
import { Route, Router } from "@solidjs/router";

import client from "@lib/client";

import SplashScreen from "@ui/splash-screen";

import Index from "./routes";

export default function App() {
	const [data, _] = createResource(() => client.queries.whoami(), { name: "whoami" });

	return (
		<>
			<Show when={data.loading}>
				<SplashScreen />
			</Show>
			<Show when={!data.loading}>
				<Router>
					<Route path="/" component={Index} />
				</Router>
			</Show>
		</>
	);
}
