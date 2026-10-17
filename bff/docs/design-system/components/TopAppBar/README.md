The top app bar names the current page and holds a few page-level actions. In the two-sided product it is used on **full-width pages without panes** (settings, client portal landing); pane-based pages use a **pane header** instead (`md-pane__head`).

**Consumer provides:** a title and up to three action icons.

- Small bar 64px; medium bar 112px with the title in `headline-medium` for friendly client pages.
- At rest it shares `surface`; once content scrolls beneath it, `surface-container`.
- Icon actions have tooltips; anything needing a label goes into a menu or button in the page.

**Implementation:** Custom Svelte component (`TopAppBar.svelte`); `@material/web` has no top app bar. Use `<md-icon-button>` for the actions.
