import { error, redirect } from '@sveltejs/kit';
import type { GetSessionResponse } from '#lib/api/session/types.js';

type Role = GetSessionResponse['role'];

const areas: Record<Role, string> = { staff: '/tax-practice', client: '/client' };

/** Redirects the actor to the home of its area, or answers 401 without one. */
export function redirectToArea(actor: GetSessionResponse | null): never {
	if (!actor) {
		// TODO(auth): redirect to the login page once there is one.
		error(401, 'Not logged in');
	}

	redirect(303, areas[actor.role]);
}

/** Lets only actors of the role through and sends everyone else to their own area. */
export function requireRole(actor: GetSessionResponse | null, role: Role): GetSessionResponse {
	if (actor?.role !== role) {
		redirectToArea(actor);
	}

	return actor;
}
