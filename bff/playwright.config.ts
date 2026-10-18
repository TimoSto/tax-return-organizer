import { defineConfig } from '@playwright/test';

export default defineConfig({
	webServer: {
		command: 'npm run build && npm run preview',
		// The e2e tests serve a mock backend on this port.
		env: { BACKEND_URL: 'http://localhost:4174' },
		port: 4173
	},
	testMatch: '**/*.e2e.{ts,js}'
});
