import { defineConfig } from '@playwright/test';

export default defineConfig({
	webServer: [
		{
			// Throwaway Postgres + API with a seeded collector (id "aaaa"); needs Go and Docker.
			command: 'go run -tags e2e ./cmd/e2e',
			cwd: '../backend',
			env: { HTTP_ADDR: ':8080' },
			url: 'http://localhost:8080/healthz',
			timeout: 180_000
		},
		{
			command: 'npm run build && npm run preview',
			env: { BACKEND_URL: 'http://localhost:8080' },
			port: 4173
		}
	],
	testMatch: '**/*.e2e.{ts,js}'
});
