<script lang="ts">
	interface Props {
		/** Material Symbols ligature name, e.g. `badge`. */
		icon: string;
	}

	let { icon }: Props = $props();
</script>

<span class="avatar"><span class="md-icon" aria-hidden="true">{icon}</span></span>

<style>
	.avatar {
		display: inline-grid;
		place-items: center;
		flex: none;
		width: 40px;
		height: 40px;
		border-radius: var(--md-sys-shape-corner-full);
		background: var(--md-sys-color-primary-container);
		color: var(--md-sys-color-on-primary-container);
	}

	:global([data-theme='high-contrast']) .avatar {
		box-shadow: inset 0 0 0 2px var(--md-sys-color-outline);
	}
</style>
