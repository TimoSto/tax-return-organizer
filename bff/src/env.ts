import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	BACKEND_URL: {
		schema: (value) => value ?? 'http://localhost:8080',
		description: 'Base URL of the backend REST API.'
	},
	// TODO: replace with the session token from the user's cookie once there is login.
	// No default: set it in .env.local (not committed) for local development.
	BACKEND_TOKEN: {
		schema: (value) => value,
		description:
			'Bearer token sent to the backend on every request. Unset sends none, so the backend answers 401.'
	}
});
