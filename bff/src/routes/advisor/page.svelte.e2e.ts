import { createServer, type Server } from 'node:http';
import { expect, test } from '@playwright/test';

// The BFF fetches server-side, so page.route() cannot intercept it. Must match BACKEND_URL in playwright.config.ts.
const MOCK_BACKEND_PORT = 4174;

let backend: Server;

test.beforeAll(async () => {
	backend = createServer((req, res) => {
		res.setHeader('Content-Type', 'application/json');

		if (req.method === 'GET' && req.url === '/collectors/00000000-0000-4000-8000-00000000aaaa') {
			res.end(
				JSON.stringify({
					id: '00000000-0000-4000-8000-00000000aaaa',
					name: 'Mock Collector',
					email: 'mock@example.com'
				})
			);

			return;
		}

		res.statusCode = 404;
		res.end(JSON.stringify({ error: 'collector not found' }));
	});

	await new Promise<void>((resolve) => backend.listen(MOCK_BACKEND_PORT, resolve));
});

test.afterAll(async () => {
	await new Promise((resolve) => backend.close(resolve));
});

test('shows the collector name', async ({ page }) => {
	await page.goto('/advisor');
	await expect(page.locator('h1')).toHaveText('Mock Collector');
});
