<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Path } from '$app/types';
	import { resolve } from '$app/paths';
	import { localizeHref } from '#lib/paraglide/runtime.js';
	import * as m from '#lib/paraglide/messages.js';
	import { isActive, type NavItem } from '#lib/nav.js';

	interface Props {
		top: NavItem[];
		bottom?: NavItem[];
		pathname: string;
		fab?: Snippet;
	}

	let { top, bottom = [], pathname, fab }: Props = $props();
</script>

{#snippet link(item: NavItem)}
	{@const active = isActive(item.href, pathname)}
	<a
		class="md-nav-item"
		class:is-active={active}
		href={resolve(localizeHref(item.href) as Path)}
		aria-current={active ? 'page' : undefined}
		aria-label={item.count ? m.nav_unread({ label: item.label(), count: item.count }) : undefined}
	>
		<span class="md-nav-item__ind">
			<span class="md-icon" aria-hidden="true">{item.icon}</span>
			{#if item.count}<span class="md-badge" aria-hidden="true">{item.count}</span>{/if}
		</span>
		{item.label()}
	</a>
{/snippet}

<nav class="md-rail" aria-label={m.nav_main()}>
	{@render fab?.()}
	<div class="md-rail__group">
		{#each top as item (item.href)}
			{@render link(item)}
		{/each}
	</div>
	<div class="md-rail__spacer"></div>
	<div class="md-rail__group">
		{#each bottom as item (item.href)}
			{@render link(item)}
		{/each}
	</div>
</nav>

<style>
	.md-nav-item {
		text-decoration: none;
	}
</style>
