import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	BACKEND_URL: {
		schema: (value) => value ?? 'http://localhost:8080',
		description: 'Base URL of the backend REST API.'
	},
	// TODO: replace with the session token from the user's cookie once there is login.
	BACKEND_TOKEN: {
		schema: (value) => value ?? 'e2e-employee-a',
		description:
			'Bearer token sent to the backend on every request. Defaults to the session of employeeA in the e2e backend (backend/cmd/e2e).'
	}
});
