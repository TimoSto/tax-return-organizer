import { isHttpError } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';
import { getTaxPractice } from './index.js';
import type { GetTaxPracticeResponse } from './types.js';

vi.mock('$app/env/private', () => ({ BACKEND_URL: 'http://backend.test' }));

function stubFetch(response: Response) {
	return vi.fn<typeof fetch>().mockResolvedValue(response);
}

describe('getTaxPractice', () => {
	it('requests the tax practice and returns the response body', async () => {
		const body: GetTaxPracticeResponse = {
			id: '00000000-0000-4000-8000-00000000aaaa',
			name: 'Steuerbüro Muster',
			email: 'info@muster.example'
		};
		const fetch = stubFetch(Response.json(body));

		await expect(getTaxPractice(fetch)).resolves.toEqual(body);
		expect(fetch).toHaveBeenCalledExactlyOnceWith('http://backend.test/tax-practice', undefined);
	});

	// The full status mapping is tested in client.spec.ts.
	it('answers 404 with a tax practice message when the backend answers 404', async () => {
		const fetch = stubFetch(new Response(null, { status: 404 }));

		await expect(getTaxPractice(fetch)).rejects.toSatisfy(
			(e) => isHttpError(e) && e.status === 404 && e.body.message === 'Tax practice not found'
		);
	});

	it('answers 401 when the backend answers 401', async () => {
		const fetch = stubFetch(new Response(null, { status: 401 }));

		await expect(getTaxPractice(fetch)).rejects.toSatisfy((e) => isHttpError(e, 401));
	});
});
