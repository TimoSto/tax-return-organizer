import { backendError, backendFetch } from '../client.js';
import type { GetSessionResponse } from './types.js';

/** Asks the backend who the caller is (GET /auth/session); null without a valid session. */
export async function getSession(
	fetch: typeof globalThis.fetch
): Promise<GetSessionResponse | null> {
	const res = await backendFetch(fetch, '/auth/session');
	if (res.status === 401) {
		return null;
	}
	if (!res.ok) {
		backendError(res);
	}

	return (await res.json()) as GetSessionResponse;
}
