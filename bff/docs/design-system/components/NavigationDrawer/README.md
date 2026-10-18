The navigation drawer carries grouped destinations. In this product it is optional for the advisor on **extra-large windows (1600px+)**, where it replaces the rail.

**Consumer provides:** grouped destinations with icon, label and optional count, and the active one.

- Width `nav-drawer-width` (360px), `surface-container-low`, trailing corners `shape-extra-large`. Items are 56px pills; active uses `secondary-container`.
- Section headings use `label-large` in `on-surface-variant`; separate sections with `md-divider`.
- Counts are numbers of pending work, never totals.

**Implementation:** Custom Svelte component (`NavigationDrawer.svelte`); `@material/web` has no drawer.
