<script lang="ts">
	import AppShell from '#lib/components/AppShell/AppShell.svelte';
	import { taxPracticeNav } from '#lib/nav.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<AppShell nav={taxPracticeNav}>{@render children()}</AppShell>
