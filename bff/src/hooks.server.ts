import { sequence } from '@sveltejs/kit/hooks';
import type { Handle, HandleFetch } from '@sveltejs/kit/hooks';
import { BACKEND_TOKEN, BACKEND_URL } from '$app/env/private';
import { getSession } from '#lib/api/session/index.js';
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
	event.locals.actor = await getSession(event.fetch);

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
