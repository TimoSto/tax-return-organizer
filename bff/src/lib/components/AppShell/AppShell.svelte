<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import { deLocalizeHref } from '#lib/paraglide/runtime.js';
	import NavigationRail from '#lib/components/NavigationRail/NavigationRail.svelte';
	import Pane from '#lib/components/Pane/Pane.svelte';
	import type { NavStructure } from '#lib/nav.js';

	interface Props {
		nav: NavStructure;
		children: Snippet;
	}

	let { nav, children }: Props = $props();

	const pathname = $derived(deLocalizeHref(page.url.pathname));
</script>

<div class="md-app shell">
	<NavigationRail top={nav.top} bottom={nav.bottom} {pathname} />

	<div class="md-panes">
		<Pane variant="detail">
			<main>{@render children()}</main>
		</Pane>
	</div>
</div>

<style>
	.shell {
		height: 100dvh;
		min-width: 1024px;
	}
</style>
