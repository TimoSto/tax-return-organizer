import { createServer, type Server } from 'node:http';
import { expect, test } from '@playwright/test';
import type { GetSessionResponse } from '#lib/api/types.js';

// The BFF fetches server-side, so page.route() cannot intercept it. Must match BACKEND_URL in playwright.config.ts.
// All e2e tests share this port, so they live in this one file: parallel workers would each start a mock on it.
const MOCK_BACKEND_PORT = 4174;
// Must match BACKEND_TOKEN in playwright.config.ts.
const MOCK_BACKEND_TOKEN = 'playwright-token';

const TAX_PRACTICE_ID = '00000000-0000-4000-8000-00000000aaaa';

const staff: GetSessionResponse = {
	role: 'staff',
	userId: '00000000-0000-4000-8000-0000000000a1',
	taxPracticeId: TAX_PRACTICE_ID
};
const client: GetSessionResponse = {
	role: 'client',
	userId: '00000000-0000-4000-8000-0000000000c1',
	taxPracticeId: TAX_PRACTICE_ID
};

// The session the mock backend resolves the token to; null answers 401 like an unknown token.
let session: GetSessionResponse | null = staff;

let backend: Server;

test.beforeAll(async () => {
	backend = createServer((req, res) => {
		res.setHeader('Content-Type', 'application/json');

		if (req.headers.authorization !== `Bearer ${MOCK_BACKEND_TOKEN}` || !session) {
			res.statusCode = 401;
			res.setHeader('WWW-Authenticate', 'Bearer');
			res.end(JSON.stringify({ code: 'unauthenticated' }));

			return;
		}

		if (req.method === 'GET' && req.url === '/auth/session') {
			res.end(JSON.stringify(session));

			return;
		}

		if (req.method === 'GET' && req.url === '/tax-practice') {
			// Like the backend for personas: the BFF must not depend on it to keep clients out of the tax practice area.
			if (session.role !== 'staff') {
				res.statusCode = 403;
				res.end(JSON.stringify({ code: 'forbidden' }));

				return;
			}

			res.end(
				JSON.stringify({
					id: TAX_PRACTICE_ID,
					name: 'Mock Tax Practice',
					email: 'mock@example.com'
				})
			);

			return;
		}

		res.statusCode = 404;
		res.end(JSON.stringify({ code: 'not_found' }));
	});

	await new Promise<void>((resolve) => backend.listen(MOCK_BACKEND_PORT, resolve));
});

test.afterAll(async () => {
	await new Promise((resolve) => backend.close(resolve));
});

// The mock's session is shared by all tests of a worker, so they must not run in parallel.
test.describe.configure({ mode: 'serial' });

test('sends staff from / to the tax practice area, which shows the tax practice name', async ({
	page
}) => {
	session = staff;
	await page.goto('/');
	await expect(page).toHaveURL('/tax-practice');
	await expect(page.locator('h1')).toHaveText('Mock Tax Practice');
});

test('sends a client from / to the client area', async ({ page }) => {
	session = client;
	await page.goto('/');
	await expect(page).toHaveURL('/client');
	await expect(page.locator('h1')).toHaveText(/Start/);
});

test('sends a client away from the tax practice area', async ({ page }) => {
	session = client;
	await page.goto('/tax-practice');
	await expect(page).toHaveURL('/client');
});

test('sends staff away from the client area', async ({ page }) => {
	session = staff;
	await page.goto('/client');
	await expect(page).toHaveURL('/tax-practice');
});

test('answers 401 without a session', async ({ page }) => {
	session = null;
	const res = await page.goto('/');
	expect(res?.status()).toBe(401);
});
