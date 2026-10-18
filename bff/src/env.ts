import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	BACKEND_URL: {
		schema: (value) => value ?? 'http://localhost:8080',
		description: 'Base URL of the backend REST API.'
	},
	// TODO: remove once the tenant is derived from the authenticated user.
	TENANT_ID: {
		schema: (value) => value ?? 'aaaa',
		description: 'Collector id used as tenantID until authentication exists.'
	}
});
