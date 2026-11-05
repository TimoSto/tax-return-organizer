<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant: 'list' | 'detail' | 'support';
		/** Accessible name of the pane, which makes it a named region. */
		label?: string;
		head?: Snippet;
		children: Snippet;
	}

	let { variant, label, head, children }: Props = $props();
</script>

<section class="md-pane md-pane--{variant}" aria-label={label}>
	{#if head}<header class="md-pane__head">{@render head()}</header>{/if}
	<div class="md-pane__scroll">{@render children()}</div>
</section>

<style>
	/* bundle.css clips panes; content panes must scroll rather than overflow. */
	.md-pane__scroll {
		overflow: auto;
		min-height: 0;
	}

	/*
	 * The scroll area clips the focus ring (3px + 2px offset, 4px in high contrast) of a control at its top, such as the
	 * search field. Give it room there and take the same space from the head, so the gap stays 12px. The detail pane's
	 * scroll area already has 16px at the top.
	 */
	.md-pane:not(.md-pane--detail) > .md-pane__head {
		padding-bottom: var(--space-1);
	}

	.md-pane:not(.md-pane--detail) > .md-pane__scroll {
		padding-top: var(--space-2);
	}
</style>
