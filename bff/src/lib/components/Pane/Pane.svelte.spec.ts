import { page } from 'vitest/browser';
import { createRawSnippet } from 'svelte';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Pane from './Pane.svelte';

function html(markup: string) {
	return createRawSnippet(() => ({ render: () => markup }));
}

describe('Pane.svelte', () => {
	it('is a region named by its label, with head and content', async () => {
		render(Pane, {
			variant: 'list',
			label: 'Personas',
			head: html('<h2>Personas</h2>'),
			children: html('<p>Inhalt</p>')
		});

		const region = page.getByRole('region', { name: 'Personas' });
		await expect.element(region.getByRole('heading', { name: 'Personas' })).toBeInTheDocument();
		await expect.element(region.getByText('Inhalt')).toBeInTheDocument();
	});

	it('is no named region and has no header without label and head', async () => {
		const { container } = render(Pane, { variant: 'detail', children: html('<p>Inhalt</p>') });

		await expect.element(page.getByText('Inhalt')).toBeInTheDocument();
		await expect.element(page.getByRole('region')).not.toBeInTheDocument();
		expect(container.querySelector('header')).toBeNull();
	});
});
