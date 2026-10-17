<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant: 'list' | 'detail' | 'support';
		head?: Snippet;
		children: Snippet;
	}

	let { variant, head, children }: Props = $props();
</script>

<section class="md-pane md-pane--{variant}">
	{#if head}<header class="md-pane__head">{@render head()}</header>{/if}
	<div class="md-pane__scroll">{@render children()}</div>
</section>

<style>
	/* bundle.css clips panes; content panes must scroll rather than overflow. */
	.md-pane__scroll {
		overflow: auto;
		min-height: 0;
	}
</style>
