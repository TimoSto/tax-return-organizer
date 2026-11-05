import { getTaxPractice } from '#lib/api/taxPractice/index.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const { name } = await getTaxPractice(fetch);

	return { name };
};
