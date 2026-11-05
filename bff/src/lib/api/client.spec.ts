import { isHttpError } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';
import { backendError, backendFetch } from './client.js';

vi.mock('$app/env/private', () => ({ BACKEND_URL: 'http://backend.test' }));

function errorOf(fn: () => unknown): { status: number; message: string } {
	try {
		fn();
	} catch (e) {
		if (isHttpError(e)) {
			return { status: e.status, message: e.body.message };
		}
		throw e;
	}
	throw new Error('expected an HttpError');
}

describe('backendFetch', () => {
	it('prefixes the path with the backend URL and passes init through', async () => {
		const res = new Response(null, { status: 204 });
		const fetch = vi.fn<typeof globalThis.fetch>().mockResolvedValue(res);
		const init: RequestInit = { method: 'POST', body: '{}' };

		await expect(backendFetch(fetch, '/tax-practice', init)).resolves.toBe(res);
		expect(fetch).toHaveBeenCalledExactlyOnceWith('http://backend.test/tax-practice', init);
	});
});

describe('backendError', () => {
	it.each([
		{ backendStatus: 401, status: 401, message: 'Not logged in' },
		{ backendStatus: 404, status: 404, message: 'Thing not found' },
		{ backendStatus: 400, status: 500, message: 'Backend request failed' },
		{ backendStatus: 403, status: 500, message: 'Backend request failed' },
		{ backendStatus: 500, status: 500, message: 'Backend request failed' },
		{ backendStatus: 503, status: 500, message: 'Backend request failed' }
	])(
		'answers $status when the backend answers $backendStatus',
		({ backendStatus, status, message }) => {
			const res = new Response(null, { status: backendStatus });

			expect(errorOf(() => backendError(res, 'Thing not found'))).toEqual({ status, message });
		}
	);

	it('has a default not found message', () => {
		expect(errorOf(() => backendError(new Response(null, { status: 404 })))).toEqual({
			status: 404,
			message: 'Not found'
		});
	});
});
