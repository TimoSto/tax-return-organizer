import { backendError, backendFetch } from '../client.js';
import type { GetTaxPracticeResponse } from './types.js';

/** Loads the tax practice of the caller (GET /tax-practice). Backend failures become SvelteKit errors. */
export async function getTaxPractice(
	fetch: typeof globalThis.fetch
): Promise<GetTaxPracticeResponse> {
	const res = await backendFetch(fetch, '/tax-practice');
	if (!res.ok) {
		backendError(res, 'Tax practice not found');
	}

	return (await res.json()) as GetTaxPracticeResponse;
}
