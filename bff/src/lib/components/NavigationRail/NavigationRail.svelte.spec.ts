import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import NavigationRail from './NavigationRail.svelte';

const items = [
	{ href: '/tax-practice', icon: 'dashboard', label: () => 'Übersicht', exact: true },
	{ href: '/tax-practice/personas', icon: 'badge', label: () => 'Personas' },
	{ href: '/tax-practice/messages', icon: 'chat_bubble', label: () => 'Nachrichten', count: 4 }
];

describe('NavigationRail.svelte', () => {
	it('marks only the current destination with aria-current', async () => {
		render(NavigationRail, { top: items, pathname: '/tax-practice/personas/42' });

		await expect
			.element(page.getByRole('link', { name: 'Personas' }))
			.toHaveAttribute('aria-current', 'page');
		await expect
			.element(page.getByRole('link', { name: 'Übersicht' }))
			.not.toHaveAttribute('aria-current');
	});

	it('gives the badge an accessible name', async () => {
		render(NavigationRail, { top: items, pathname: '/tax-practice' });

		await expect
			.element(page.getByRole('link', { name: /4 ungelesen|4 unread/ }))
			.toBeInTheDocument();
	});
});
