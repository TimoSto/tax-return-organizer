import { error } from '@sveltejs/kit';
import { BACKEND_URL } from '$app/env/private';

/**
 * Sends a request to the backend. Pass SvelteKit's `fetch`, so `handleFetch` authenticates the request. `path` starts
 * with a slash, e.g. `/tax-practice`.
 */
export function backendFetch(
	fetch: typeof globalThis.fetch,
	path: string,
	init?: RequestInit
): Promise<Response> {
	return fetch(`${BACKEND_URL}${path}`, init);
}

/** Turns a failed backend response into a SvelteKit error: 401 and 404 pass through, everything else becomes 500. */
export function backendError(res: Response, notFoundMessage = 'Not found'): never {
	if (res.status === 401) {
		error(401, 'Not logged in');
	}
	if (res.status === 404) {
		error(404, notFoundMessage);
	}
	error(500, 'Backend request failed');
}
