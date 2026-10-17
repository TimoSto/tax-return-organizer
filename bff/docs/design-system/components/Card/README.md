Cards group related content about one subject: a figure, a deadline, a summary.

**Consumer provides:** content with a `title-small` heading or a `label-medium` eyebrow. Cards are not links by themselves; add a button, or make the whole card one state-layer target, never both.

- Filled (default) for most content; outlined when the card sits on a filled surface; elevated only for floating or draggable cards.
- Tertiary is reserved for deadlines and "action required"; primary for the single most important result.
- Radius `shape-medium`, padding `space-4`. Do not nest cards.
- In high contrast every card has a 2px `outline` border.

**Implementation:** Custom Svelte component (`Card.svelte`) styled with the `md-card` rules; `@material/web` has no card.
