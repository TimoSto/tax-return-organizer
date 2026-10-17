<script lang="ts">
	import type { Path } from '$app/types';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { deLocalizeHref, locales, localizeHref } from '#lib/paraglide/runtime.js';
	import '#lib/design-system/tokens.css';
	import '#lib/design-system/bundle.css';
	import NavigationRail from '#lib/components/NavigationRail/NavigationRail.svelte';
	import NavigationDrawer from '#lib/components/NavigationDrawer/NavigationDrawer.svelte';
	import Pane from '#lib/components/Pane/Pane.svelte';
	import { navFor, settingsItem, type Role } from '#lib/nav.js';
	import favicon from '#lib/assets/favicon.svg';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	// TODO: derive from the authenticated user once roles exist.
	const role: Role = 'advisor';

	// `--window-extra-large` in tokens.css
	const extraLarge = '(min-width: 1600px)';
	let isExtraLarge = $state(false);

	$effect(() => {
		const query = matchMedia(extraLarge);
		isExtraLarge = query.matches;
		const onChange = (e: MediaQueryListEvent) => (isExtraLarge = e.matches);
		query.addEventListener('change', onChange);
		return () => query.removeEventListener('change', onChange);
	});

	const groups = $derived(navFor(role));
	const pathname = $derived(deLocalizeHref(page.url.pathname));
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="md-app shell">
	{#if role === 'advisor' && isExtraLarge}
		<NavigationDrawer {groups} footer={settingsItem} {pathname} />
	{:else}
		<NavigationRail items={groups.flatMap((g) => g.items)} footer={settingsItem} {pathname} />
	{/if}

	<div class="md-panes">
		<Pane variant="detail">
			<main>{@render children()}</main>
		</Pane>
	</div>
</div>

<div style="display:none">
	{#each locales as locale (locale)}
		<a href={resolve(localizeHref(page.url.pathname, { locale }) as Path)}>{locale}</a>
	{/each}
</div>

<style>
	:global(html),
	:global(body) {
		margin: 0;
		height: 100%;
		background: var(--surface);
		color: var(--on-surface);
		font-family: var(--font-sans);
	}

	.shell {
		height: 100dvh;
		min-width: 1024px;
	}

	@media (prefers-reduced-motion: reduce) {
		:global(*),
		:global(*::before),
		:global(*::after) {
			transition-duration: 0s !important;
			animation-duration: 0s !important;
		}
	}
</style>
