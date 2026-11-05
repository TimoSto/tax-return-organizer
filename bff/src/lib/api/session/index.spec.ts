import { isHttpError } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';
import { getSession } from './index.js';
import type { GetSessionResponse } from './types.js';

vi.mock('$app/env/private', () => ({ BACKEND_URL: 'http://backend.test' }));

function stubFetch(response: Response) {
	return vi.fn<typeof fetch>().mockResolvedValue(response);
}

describe('getSession', () => {
	it('requests the session and returns the response body', async () => {
		const body: GetSessionResponse = {
			role: 'staff',
			userId: '00000000-0000-4000-8000-000000000001',
			taxPracticeId: '00000000-0000-4000-8000-00000000aaaa'
		};
		const fetch = stubFetch(Response.json(body));

		await expect(getSession(fetch)).resolves.toEqual(body);
		expect(fetch).toHaveBeenCalledExactlyOnceWith('http://backend.test/auth/session', undefined);
	});

	it('returns null when the backend answers 401', async () => {
		const fetch = stubFetch(new Response(null, { status: 401 }));

		await expect(getSession(fetch)).resolves.toBeNull();
	});

	it('answers 500 when the backend fails', async () => {
		const fetch = stubFetch(new Response(null, { status: 503 }));

		await expect(getSession(fetch)).rejects.toSatisfy((e) => isHttpError(e) && e.status === 500);
	});
});
