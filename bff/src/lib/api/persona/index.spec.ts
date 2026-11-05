import { isHttpError } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';
import { getPersona, listPersonas } from './index.js';
import type { GetPersonaResponse } from './types.js';

vi.mock('$app/env/private', () => ({ BACKEND_URL: 'http://backend.test' }));

function stubFetch(response: Response) {
	return vi.fn<typeof fetch>().mockResolvedValue(response);
}

function isError(status: number, message: string) {
	return (e: unknown) => isHttpError(e, status) && e.body.message === message;
}

const persona: GetPersonaResponse = {
	id: '00000000-0000-4000-8000-000000000101',
	taxPracticeId: '00000000-0000-4000-8000-00000000aaaa',
	name: 'Employed',
	description: 'Has a job',
	items: [
		{
			id: '00000000-0000-4000-8000-000000000201',
			title: 'Payslip',
			kind: 'document',
			recurrence: 'monthly',
			documentSpec: {
				metadataFields: [
					{ name: 'Employer', valueSpec: { type: 'text', unit: '', required: true } }
				]
			},
			valueSpec: null,
			instanceFields: []
		}
	]
};

describe('listPersonas', () => {
	it('requests the personas and returns them unwrapped', async () => {
		const fetch = stubFetch(Response.json({ personas: [persona] }));

		await expect(listPersonas(fetch)).resolves.toEqual([persona]);
		expect(fetch).toHaveBeenCalledExactlyOnceWith('http://backend.test/personas', undefined);
	});

	it('returns an empty list when the tax practice has no personas', async () => {
		const fetch = stubFetch(Response.json({ personas: [] }));

		await expect(listPersonas(fetch)).resolves.toEqual([]);
	});

	// The full status mapping is tested in client.spec.ts.
	it('answers 404 with a tax practice message when the backend answers 404', async () => {
		const fetch = stubFetch(new Response(null, { status: 404 }));

		await expect(listPersonas(fetch)).rejects.toSatisfy(isError(404, 'Tax practice not found'));
	});
});

describe('getPersona', () => {
	it('requests the persona and returns the response body', async () => {
		const fetch = stubFetch(Response.json(persona));

		await expect(getPersona(fetch, persona.id)).resolves.toEqual(persona);
		expect(fetch).toHaveBeenCalledExactlyOnceWith(
			`http://backend.test/personas/${persona.id}`,
			undefined
		);
	});

	it('encodes the id into the path', async () => {
		const fetch = stubFetch(Response.json(persona));

		await getPersona(fetch, '../tax-practice?x=1');

		expect(fetch).toHaveBeenCalledExactlyOnceWith(
			'http://backend.test/personas/..%2Ftax-practice%3Fx%3D1',
			undefined
		);
	});

	it.each([400, 404])('answers 404 when the backend answers %i', async (backendStatus) => {
		const fetch = stubFetch(new Response(null, { status: backendStatus }));

		await expect(getPersona(fetch, persona.id)).rejects.toSatisfy(
			isError(404, 'Persona not found')
		);
	});

	it('answers 401 when the backend answers 401', async () => {
		const fetch = stubFetch(new Response(null, { status: 401 }));

		await expect(getPersona(fetch, persona.id)).rejects.toSatisfy((e) => isHttpError(e, 401));
	});
});
