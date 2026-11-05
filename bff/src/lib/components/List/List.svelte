<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		'aria-label'?: string;
		children: Snippet;
	}

	let { children, ...rest }: Props = $props();
</script>

<ul class="list" {...rest}>{@render children()}</ul>

<style>
	.list {
		display: flex;
		flex-direction: column;
		margin: 0;
		padding: 0;
		list-style: none;
	}
</style>
