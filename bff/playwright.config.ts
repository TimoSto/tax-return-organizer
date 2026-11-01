import { defineConfig } from '@playwright/test';

export default defineConfig({
	webServer: {
		command: 'npm run build && npm run preview',
		// The e2e tests serve a mock backend on this port, which only accepts this token.
		env: { BACKEND_URL: 'http://localhost:4174', BACKEND_TOKEN: 'playwright-token' },
		port: 4173
	},
	testMatch: '**/*.e2e.{ts,js}'
});
