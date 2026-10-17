import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import * as m from '#lib/paraglide/messages.js';
import NavigationDrawer from './NavigationDrawer.svelte';

const groups = [
	{
		heading: () => 'Kanzlei',
		items: [
			{ href: '/', icon: 'dashboard', label: () => 'Übersicht' },
			{ href: '/mandanten', icon: 'group', label: () => 'Mandanten' }
		]
	},
	{
		heading: () => 'Zusammenarbeit',
		items: [{ href: '/nachrichten', icon: 'chat_bubble', label: () => 'Nachrichten', count: 4 }]
	}
];

const footer = { href: '/einstellungen', icon: 'settings', label: () => 'Einstellungen' };

describe('NavigationDrawer.svelte', () => {
	it('renders a labelled navigation landmark with group headings', async () => {
		render(NavigationDrawer, { groups, pathname: '/' });

		await expect.element(page.getByRole('navigation', { name: m.nav_main() })).toBeInTheDocument();
		await expect.element(page.getByText('Kanzlei')).toBeInTheDocument();
		await expect.element(page.getByText('Zusammenarbeit')).toBeInTheDocument();
	});

	it('marks only the current destination with aria-current', async () => {
		render(NavigationDrawer, { groups, pathname: '/mandanten/42' });

		await expect
			.element(page.getByRole('link', { name: 'Mandanten' }))
			.toHaveAttribute('aria-current', 'page');
		await expect
			.element(page.getByRole('link', { name: 'Übersicht' }))
			.not.toHaveAttribute('aria-current');
	});

	it('gives the count badge an accessible name', async () => {
		render(NavigationDrawer, { groups, pathname: '/' });

		await expect
			.element(page.getByRole('link', { name: m.nav_pending({ label: 'Nachrichten', count: 4 }) }))
			.toBeInTheDocument();
	});

	it('renders no badge when there is no count', async () => {
		render(NavigationDrawer, { groups, pathname: '/' });

		await expect
			.element(page.getByRole('link', { name: 'Übersicht' }))
			.not.toHaveAttribute('aria-label');
	});

	it('renders the footer item only when provided', async () => {
		const { rerender } = render(NavigationDrawer, { groups, pathname: '/' });

		await expect.element(page.getByRole('link', { name: 'Einstellungen' })).not.toBeInTheDocument();

		await rerender({ groups, footer, pathname: '/einstellungen' });

		await expect
			.element(page.getByRole('link', { name: 'Einstellungen' }))
			.toHaveAttribute('aria-current', 'page');
	});
});
