The product is **desktop only**: the supported minimum width is **1024px**. Layout adapts to window width using the three Material Design 3 desktop window size classes. Compact and medium classes (phones, tablets) are out of scope; do not build bottom navigation bars or single-pane back stacks.

| Class | Width | Navigation | Advisor panes | Client panes |
|---|---|---|---|---|
| Expanded | 840–1199 | Rail | List + detail; supporting pane as side sheet | Detail; supporting pane below the content |
| Large | 1200–1599 | Rail | List + detail + supporting | Detail + supporting |
| Extra-large | 1600+ | Rail or drawer | Same; detail grows | Same; content max 1040px |

- Use the `window-*` tokens in `@media (min-width: …)` queries. Fixed panes (`pane-list-width`, `pane-supporting-width`) keep their size; only the detail pane flexes.
- Margin and gutter are `pane-margin` / `pane-gutter` (24px) in every class.
- Test at 1024, 1280, 1440, 1920 and with browser zoom at 200% (which behaves like a 960px window; the layout must still work as the expanded class with the supporting pane as a sheet).

**Implementation:** A Svelte store or `matchMedia` on the `window-*` values decides how many panes render; panes themselves are CSS grid columns.
