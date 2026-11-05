import { listPersonas } from '#lib/api/persona/index.js';
import { requireRole } from '#lib/server/area.js';
import type { LayoutServerLoad } from './$types';

// Page loads run in parallel with this one and may still ask the backend, which then answers 403 for the wrong
// role. The redirect wins all the same, because SvelteKit handles the errors of the outer layouts first.
export const load: LayoutServerLoad = async ({ locals, fetch }) => {
	requireRole(locals.actor, 'staff');

	const personas = await listPersonas(fetch);

	// The backend returns them in no particular order.
	return { personas: personas.toSorted((a, b) => a.name.localeCompare(b.name)) };
};
