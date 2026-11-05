<script lang="ts">
	import type { GetPersonaResponse } from '#lib/api/persona/types.js';
	import * as m from '#lib/paraglide/messages.js';
	import Avatar from '#lib/components/Avatar/Avatar.svelte';
	import List from '#lib/components/List/List.svelte';
	import ListItem from '#lib/components/List/ListItem.svelte';
	import SearchField from '#lib/components/SearchField/SearchField.svelte';

	interface Props {
		personas: GetPersonaResponse[];
	}

	let { personas }: Props = $props();

	let query = $state('');

	const visible = $derived.by(() => {
		const q = query.trim().toLocaleLowerCase();
		return q ? personas.filter((p) => p.name.toLocaleLowerCase().includes(q)) : personas;
	});

	function summary(persona: GetPersonaResponse): string {
		const documents = persona.items.filter((i) => i.kind === 'document').length;
		const data = persona.items.length - documents;
		return m.persona_summary({
			documents: m.persona_documents({ count: documents }),
			data: m.persona_data({ count: data })
		});
	}
</script>

<div class="persona-list">
	{#if personas.length === 0}
		<p class="empty">{m.personas_empty()}</p>
	{:else}
		<SearchField bind:value={query} label={m.personas_search()} />

		{#if visible.length > 0}
			<List aria-label={m.nav_personas()}>
				{#each visible as persona (persona.id)}
					<ListItem headline={persona.name} supporting={summary(persona)}>
						{#snippet leading()}<Avatar icon="badge" />{/snippet}
					</ListItem>
				{/each}
			</List>
		{:else}
			<p class="empty">{m.personas_no_match()}</p>
		{/if}
	{/if}
</div>

<style>
	.persona-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.empty {
		margin: 0;
		padding: var(--space-2) var(--space-4);
		color: var(--md-sys-color-on-surface-variant);
		font-family: var(--md-sys-typescale-body-medium-font);
		font-size: var(--md-sys-typescale-body-medium-size);
		line-height: var(--md-sys-typescale-body-medium-line-height);
		letter-spacing: var(--md-sys-typescale-body-medium-tracking);
	}
</style>
