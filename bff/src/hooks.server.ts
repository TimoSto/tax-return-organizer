import type { Handle } from '@sveltejs/kit/hooks';
import { sequence } from '@sveltejs/kit/hooks';
import { TENANT_ID } from '$app/env/private';
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

// TODO: derive the tenant from the authenticated user once authentication exists.
const handleTenant: Handle = ({ event, resolve }) => {
	event.locals.tenantID = TENANT_ID;

	return resolve(event);
};

export const handle: Handle = sequence(handleTenant, handleParaglide);
