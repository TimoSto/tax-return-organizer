<script lang="ts">
	import type { Path } from '$app/types';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { locales, localizeHref } from '#lib/paraglide/runtime.js';
	import '@fontsource/roboto/400.css';
	import '@fontsource/roboto/500.css';
	import '@fontsource/roboto/700.css';
	import '#lib/design-system/tokens.css';
	import '#lib/design-system/bundle.css';
	import favicon from '#lib/assets/favicon.svg';
	import appleTouchIcon from '#lib/assets/apple-touch-icon.png';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} type="image/svg+xml" />
	<link rel="apple-touch-icon" href={appleTouchIcon} />
</svelte:head>

{@render children()}

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

	@media (prefers-reduced-motion: reduce) {
		:global(*),
		:global(*::before),
		:global(*::after) {
			transition-duration: 0s !important;
			animation-duration: 0s !important;
		}
	}
</style>
