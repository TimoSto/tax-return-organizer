import { error } from '@sveltejs/kit';
import { BACKEND_URL } from '$app/env/private';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch }) => {
	const res = await fetch(`${BACKEND_URL}/collectors/${encodeURIComponent(locals.tenantID)}`);

	if (res.status === 404) {
		error(404, 'Collector not found');
	}
	if (!res.ok) {
		error(502, 'Backend request failed');
	}

	const { name } = (await res.json()) as { name: string };

	return { name };
};
