<script lang="ts">
	interface Props {
		value: string;
		/** Shown as placeholder and used as the accessible name. */
		label: string;
	}

	let { value = $bindable(), label }: Props = $props();
</script>

<!-- Material Web has no search bar; this follows .md-search in docs/design-system/components/bundle.css. -->
<label class="search">
	<span class="md-icon" aria-hidden="true">search</span>
	<input type="search" bind:value placeholder={label} aria-label={label} />
</label>

<style>
	.search {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		height: 56px;
		padding: 0 var(--space-4);
		box-sizing: border-box;
		border-radius: var(--md-sys-shape-corner-full);
		background: var(--md-sys-color-surface-container-high);
		color: var(--md-sys-color-on-surface-variant);
		cursor: text;
	}

	.search:focus-within {
		outline: 3px solid var(--focus-ring);
		outline-offset: 2px;
	}

	input {
		flex: 1;
		min-width: 0;
		border: none;
		outline: none;
		padding: 0;
		background: transparent;
		color: var(--md-sys-color-on-surface);
		font-family: var(--md-sys-typescale-body-large-font);
		font-size: var(--md-sys-typescale-body-large-size);
		line-height: var(--md-sys-typescale-body-large-line-height);
		font-weight: var(--md-sys-typescale-body-large-weight);
		letter-spacing: var(--md-sys-typescale-body-large-tracking);
	}

	input::placeholder {
		color: var(--md-sys-color-on-surface-variant);
		opacity: 1;
	}

	:global([data-theme='high-contrast']) .search {
		box-shadow: inset 0 0 0 2px var(--md-sys-color-outline);
	}

	:global([data-theme='high-contrast']) .search:focus-within {
		outline-width: 4px;
	}
</style>
