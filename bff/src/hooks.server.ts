import { error } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import type { Handle, HandleFetch } from '@sveltejs/kit/hooks';
import { BACKEND_TOKEN, BACKEND_URL } from '$app/env/private';
import type { GetSessionResponse } from '#lib/api/types.js';
import { getTextDirection } from '#lib/paraglide/runtime.js';
import { paraglideMiddleware } from '#lib/paraglide/server.js';

const handleParaglide: Handle = ({ event, resolve }) =>
	paraglideMiddleware(event.request, ({ request, locale }) => {
		return resolve(
			{ ...event, request },
			{
				transformPageChunk: ({ html }) =>
					html
						.replace('%paraglide.lang%', locale)
						.replace('%paraglide.dir%', getTextDirection(locale))
			}
		);
	});

// Asks the backend who the caller is, so layouts can pick the area (tax practice or client) without each page asking.
// It decides only what to show: the backend still checks every request itself.
const handleAuth: Handle = async ({ event, resolve }) => {
	const res = await event.fetch(`${BACKEND_URL}/auth/session`);

	if (res.status === 401) {
		event.locals.actor = null;
	} else if (res.ok) {
		event.locals.actor = (await res.json()) as GetSessionResponse;
	} else {
		error(500, 'Backend request failed');
	}

	return resolve(event);
};

export const handle: Handle = sequence(handleParaglide, handleAuth);

// Authenticates every request to the backend, so no load function has to. The backend takes the tenant from the
// session, so requests never name it.
export const handleFetch: HandleFetch = ({ request, fetch }) => {
	if (BACKEND_TOKEN && request.url.startsWith(`${BACKEND_URL}/`)) {
		request.headers.set('Authorization', `Bearer ${BACKEND_TOKEN}`);
	}

	return fetch(request);
};
