import { redirectToArea } from '#lib/server/area.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ locals }) => redirectToArea(locals.actor);
