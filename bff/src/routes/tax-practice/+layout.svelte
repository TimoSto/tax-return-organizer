<script lang="ts">
	import AppShell from '#lib/components/AppShell/AppShell.svelte';
	import Pane from '#lib/components/Pane/Pane.svelte';
	import PersonaList from '#lib/components/PersonaList/PersonaList.svelte';
	import * as m from '#lib/paraglide/messages.js';
	import { taxPracticeNav } from '#lib/nav.js';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
</script>

<AppShell nav={taxPracticeNav}>
	{#snippet list()}
		<Pane variant="list" label={m.nav_personas()}>
			{#snippet head()}<h2 class="title">{m.nav_personas()}</h2>{/snippet}
			<PersonaList personas={data.personas} />
		</Pane>
	{/snippet}

	{@render children()}
</AppShell>

<style>
	.title {
		margin: 0;
		font-family: var(--md-sys-typescale-title-large-font);
		font-size: var(--md-sys-typescale-title-large-size);
		line-height: var(--md-sys-typescale-title-large-line-height);
		font-weight: var(--md-sys-typescale-title-large-weight);
		letter-spacing: var(--md-sys-typescale-title-large-tracking);
	}
</style>
