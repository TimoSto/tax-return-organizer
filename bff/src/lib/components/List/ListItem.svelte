<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		headline: string;
		supporting?: string;
		leading?: Snippet;
		trailing?: Snippet;
	}

	let { headline, supporting, leading, trailing }: Props = $props();
</script>

<!-- Mirrors .md-list__* in docs/design-system/components/bundle.css. Not interactive, so no state layer. -->
<li class="item" class:two-line={supporting}>
	{@render leading?.()}
	<div class="body">
		<div class="headline">{headline}</div>
		{#if supporting}<div class="supporting">{supporting}</div>{/if}
	</div>
	{@render trailing?.()}
</li>

<style>
	.item {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		min-height: 56px;
		padding: var(--space-2) var(--space-4);
		box-sizing: border-box;
		border-radius: var(--md-sys-shape-corner-large);
		color: var(--md-sys-color-on-surface);
	}

	.two-line {
		min-height: 72px;
	}

	.body {
		flex: 1;
		min-width: 0;
	}

	.headline,
	.supporting {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.headline {
		font-family: var(--md-sys-typescale-title-medium-font);
		font-size: var(--md-sys-typescale-title-medium-size);
		line-height: var(--md-sys-typescale-title-medium-line-height);
		font-weight: var(--md-sys-typescale-title-medium-weight);
		letter-spacing: var(--md-sys-typescale-title-medium-tracking);
	}

	.supporting {
		font-family: var(--md-sys-typescale-body-medium-font);
		font-size: var(--md-sys-typescale-body-medium-size);
		line-height: var(--md-sys-typescale-body-medium-line-height);
		font-weight: var(--md-sys-typescale-body-medium-weight);
		letter-spacing: var(--md-sys-typescale-body-medium-tracking);
		color: var(--md-sys-color-on-surface-variant);
	}
</style>
