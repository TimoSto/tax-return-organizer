import { error } from '@sveltejs/kit';
import { backendError, backendFetch } from '../client.js';
import type { GetPersonaResponse, ListPersonasResponse } from './types.js';

/** Lists the personas of the caller's tax practice (GET /personas), in no particular order. */
export async function listPersonas(fetch: typeof globalThis.fetch): Promise<GetPersonaResponse[]> {
	const res = await backendFetch(fetch, '/personas');
	if (!res.ok) {
		backendError(res, 'Tax practice not found');
	}

	const { personas } = (await res.json()) as ListPersonasResponse;
	return personas;
}

/** Loads a persona of the caller's tax practice (GET /personas/{id}). */
export async function getPersona(
	fetch: typeof globalThis.fetch,
	id: string
): Promise<GetPersonaResponse> {
	const res = await backendFetch(fetch, `/personas/${encodeURIComponent(id)}`);
	// The backend answers 400 for a malformed id; for the user that is a persona that does not exist.
	if (res.status === 400) {
		error(404, 'Persona not found');
	}
	if (!res.ok) {
		backendError(res, 'Persona not found');
	}

	return (await res.json()) as GetPersonaResponse;
}
