import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import SearchField from './SearchField.svelte';

describe('SearchField.svelte', () => {
	it('is a search box named and labelled by the label', async () => {
		render(SearchField, { value: '', label: 'Personas durchsuchen' });

		const box = page.getByRole('searchbox', { name: 'Personas durchsuchen' });
		await expect.element(box).toBeInTheDocument();
		await expect.element(box).toHaveAttribute('placeholder', 'Personas durchsuchen');
	});

	it('shows the given value', async () => {
		render(SearchField, { value: 'Verm', label: 'Suche' });

		await expect.element(page.getByRole('searchbox')).toHaveValue('Verm');
	});

	it('focuses the input when the field is clicked outside the input', async () => {
		render(SearchField, { value: '', label: 'Suche' });

		// The icon is part of the label, so clicking it focuses the input.
		await page.getByText('search').click();

		await expect.element(page.getByRole('searchbox')).toHaveFocus();
	});
});
