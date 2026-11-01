import { error } from '@sveltejs/kit';
import { BACKEND_URL } from '$app/env/private';
import type { GetCollectorResponse } from '#lib/api/types.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const res = await fetch(`${BACKEND_URL}/collector`);

	if (res.status === 401) {
		error(401, 'Not logged in');
	}

	if (res.status === 404) {
		error(404, 'Collector not found');
	}
	if (!res.ok) {
		error(500, 'Backend request failed');
	}

	const { name } = (await res.json()) as GetCollectorResponse;

	return { name };
};
