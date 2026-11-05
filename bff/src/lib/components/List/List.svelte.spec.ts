import { page } from 'vitest/browser';
import { createRawSnippet } from 'svelte';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import List from './List.svelte';
import ListItem from './ListItem.svelte';

function html(markup: string) {
	return createRawSnippet(() => ({ render: () => markup }));
}

describe('List.svelte', () => {
	it('renders a list with its accessible name and children', async () => {
		render(List, { 'aria-label': 'Personas', children: html('<li>Angestellt</li>') });

		const list = page.getByRole('list', { name: 'Personas' });
		await expect.element(list).toBeInTheDocument();
		await expect.element(list.getByRole('listitem')).toHaveTextContent('Angestellt');
	});
});

describe('ListItem.svelte', () => {
	it('renders headline and supporting text', async () => {
		render(ListItem, { headline: 'Angestellt', supporting: '6 Unterlagen · 3 Angaben' });

		await expect.element(page.getByText('Angestellt')).toBeInTheDocument();
		await expect.element(page.getByText('6 Unterlagen · 3 Angaben')).toBeInTheDocument();
	});

	it('renders only the headline without supporting text', async () => {
		render(ListItem, { headline: 'Angestellt' });

		await expect.element(page.getByRole('listitem')).toHaveTextContent(/^\s*Angestellt\s*$/);
	});

	it('renders leading and trailing content around the text', async () => {
		render(ListItem, {
			headline: 'Angestellt',
			leading: html('<span>lead</span>'),
			trailing: html('<span>trail</span>')
		});

		await expect.element(page.getByRole('listitem')).toHaveTextContent(/lead\s*Angestellt\s*trail/);
	});
});
