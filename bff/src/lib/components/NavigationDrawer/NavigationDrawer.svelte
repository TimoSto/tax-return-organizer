<script lang="ts">
	import type { Path } from '$app/types';
	import { resolve } from '$app/paths';
	import { localizeHref } from '#lib/paraglide/runtime.js';
	import * as m from '#lib/paraglide/messages.js';
	import { isActive, type NavGroup, type NavItem } from '#lib/nav.js';

	interface Props {
		groups: NavGroup[];
		footer?: NavItem;
		pathname: string;
	}

	let { groups, footer, pathname }: Props = $props();
</script>

{#snippet link(item: NavItem)}
	{@const active = isActive(item.href, pathname)}
	<a
		class="md-drawer__item md-state"
		class:is-active={active}
		href={resolve(localizeHref(item.href) as Path)}
		aria-current={active ? 'page' : undefined}
		aria-label={item.count ? m.nav_pending({ label: item.label(), count: item.count }) : undefined}
	>
		<span class="md-icon" aria-hidden="true">{item.icon}</span>
		<span>{item.label()}</span>
		{#if item.count}<span class="md-drawer__count" aria-hidden="true">{item.count}</span>{/if}
	</a>
{/snippet}

<nav class="md-drawer" aria-label={m.nav_main()}>
	{#each groups as group, i (i)}
		{#if i > 0}<hr class="md-divider" />{/if}
		<div class="md-drawer__head">{group.heading()}</div>
		{#each group.items as item (item.href)}
			{@render link(item)}
		{/each}
	{/each}
	<div class="spacer"></div>
	{#if footer}{@render link(footer)}{/if}
</nav>

<style>
	.md-drawer {
		flex: none;
	}
	.md-drawer__item {
		text-decoration: none;
	}
	.spacer {
		flex: 1;
	}
</style>
