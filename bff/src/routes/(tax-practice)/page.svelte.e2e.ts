import { createServer, type Server } from 'node:http';
import { expect, test } from '@playwright/test';

// The BFF fetches server-side, so page.route() cannot intercept it. Must match BACKEND_URL in playwright.config.ts.
const MOCK_BACKEND_PORT = 4174;
// Must match BACKEND_TOKEN in playwright.config.ts.
const MOCK_BACKEND_TOKEN = 'playwright-token';

let backend: Server;

test.beforeAll(async () => {
	backend = createServer((req, res) => {
		res.setHeader('Content-Type', 'application/json');

		if (req.headers.authorization !== `Bearer ${MOCK_BACKEND_TOKEN}`) {
			res.statusCode = 401;
			res.setHeader('WWW-Authenticate', 'Bearer');
			res.end(JSON.stringify({ code: 'unauthenticated' }));

			return;
		}

		if (req.method === 'GET' && req.url === '/tax-practice') {
			res.end(
				JSON.stringify({
					id: '00000000-0000-4000-8000-00000000aaaa',
					name: 'Mock Tax Practice',
					email: 'mock@example.com'
				})
			);

			return;
		}

		res.statusCode = 404;
		res.end(JSON.stringify({ error: 'tax practice not found' }));
	});

	await new Promise<void>((resolve) => backend.listen(MOCK_BACKEND_PORT, resolve));
});

test.afterAll(async () => {
	await new Promise((resolve) => backend.close(resolve));
});

test('shows the tax practice name', async ({ page }) => {
	await page.goto('/');
	await expect(page.locator('h1')).toHaveText('Mock Tax Practice');
});
