import path from "node:path";
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

export default defineConfig({
	plugins: [solid()],
	resolve: {
		alias: {
			$: path.resolve(__dirname, "src"),
			"@ui": path.resolve(__dirname, "src/components"),
			"@lib": path.resolve(__dirname, "src/lib"),
		},
	},
});
