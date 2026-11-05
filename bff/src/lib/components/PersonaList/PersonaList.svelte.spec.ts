import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { GetPersonaResponse, ItemDefinition } from '#lib/api/persona/types.js';
import PersonaList from './PersonaList.svelte';

function item(kind: ItemDefinition['kind']): ItemDefinition {
	return {
		id: crypto.randomUUID(),
		title: kind,
		kind,
		recurrence: 'annually',
		documentSpec: kind === 'document' ? { metadataFields: [] } : null,
		valueSpec: kind === 'data' ? { type: 'text', unit: '', required: true } : null,
		instanceFields: []
	};
}

function persona(name: string, items: ItemDefinition[] = []): GetPersonaResponse {
	return {
		id: crypto.randomUUID(),
		taxPracticeId: '00000000-0000-4000-8000-00000000aaaa',
		name,
		description: '',
		items
	};
}

// The test runs in either locale, so texts match both.
const personas = [
	persona('Angestellt', [item('document'), item('data'), item('data')]),
	persona('Vermietung', [item('document'), item('document'), item('data')])
];

describe('PersonaList.svelte', () => {
	it('lists the personas with their document and data counts', async () => {
		render(PersonaList, { personas });

		const items = page.getByRole('listitem');
		await expect.element(items.nth(0)).toHaveTextContent('Angestellt');
		await expect
			.element(items.nth(0))
			.toHaveTextContent(/1 (Unterlage|document) · 2 (Angaben|fields)/);
		await expect
			.element(items.nth(1))
			.toHaveTextContent(/2 (Unterlagen|documents) · 1 (Angabe|field)\b/);
	});

	it('filters the personas by name, ignoring case', async () => {
		render(PersonaList, { personas });

		await page.getByRole('searchbox').fill('VERM');

		await expect.element(page.getByRole('listitem')).toHaveTextContent('Vermietung');
		expect(page.getByRole('listitem').all()).toHaveLength(1);
	});

	it('says so when the search matches no persona', async () => {
		render(PersonaList, { personas });

		await page.getByRole('searchbox').fill('xyz');

		await expect
			.element(page.getByText(/Keine Persona gefunden|No persona found/))
			.toBeInTheDocument();
		await expect.element(page.getByRole('list')).not.toBeInTheDocument();
	});

	it('says so when there are no personas', async () => {
		render(PersonaList, { personas: [] });

		await expect.element(page.getByText(/Noch keine Personas|No personas yet/)).toBeInTheDocument();
		await expect.element(page.getByRole('searchbox')).not.toBeInTheDocument();
	});
});
